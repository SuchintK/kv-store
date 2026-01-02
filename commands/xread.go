package command

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type XReadCommand Command

func (cmd *XReadCommand) Execute(con *client.Client) RESPValue {
	// Minimum args: STREAMS key id (3 args after command name)
	if len(cmd.args) < 3 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	// Parse BLOCK option if present
	var blockTimeout int64 = -1 // -1 means no blocking
	argsToProcess := cmd.args

	// Check for BLOCK keyword
	if len(cmd.args) >= 2 && strings.ToUpper(cmd.args[0]) == "BLOCK" {
		timeout, err := strconv.ParseInt(cmd.args[1], 10, 64)
		if err != nil {
			return resp.EncodeSimpleError(errSyntax)
		}
		blockTimeout = timeout
		argsToProcess = cmd.args[2:] // Skip BLOCK and timeout
	}

	// Find STREAMS keyword
	streamsIdx := -1
	for i, arg := range argsToProcess {
		if strings.ToUpper(arg) == "STREAMS" {
			streamsIdx = i
			break
		}
	}

	if streamsIdx == -1 {
		return resp.EncodeSimpleError(errSyntax)
	}

	// Args after STREAMS should be: key1 key2 ... keyN id1 id2 ... idN
	argsAfterStreams := argsToProcess[streamsIdx+1:]
	if len(argsAfterStreams) < 2 || len(argsAfterStreams)%2 != 0 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	numStreams := len(argsAfterStreams) / 2
	keys := argsAfterStreams[:numStreams]
	ids := argsAfterStreams[numStreams:]

	// If using $, replace with actual last IDs for blocking
	resolvedIDs := make([]string, len(ids))
	for i, id := range ids {
		if id == "$" {
			val, exists := store.Get(keys[i])
			if exists && val.StreamData != nil && len(val.StreamData.Entries) > 0 {
				resolvedIDs[i] = val.StreamData.LastID
			} else {
				resolvedIDs[i] = "0-0"
			}
		} else {
			resolvedIDs[i] = id
		}
	}

	// Try to read entries
	results := readStreams(keys, resolvedIDs)

	// If blocking and no results, wait for new entries
	if blockTimeout >= 0 && len(results) == 0 {
		return blockAndWait(keys, resolvedIDs, blockTimeout)
	}

	// Return null if no entries found in any stream
	if len(results) == 0 {
		return resp.EncodeNullBulkString()
	}

	return resp.EncodeArray(results)
}

// readStreams reads entries from multiple streams
func readStreams(keys []string, ids []string) [][]byte {
	var results [][]byte
	for i := 0; i < len(keys); i++ {
		streamKey := keys[i]
		startID := ids[i]

		// Get stream
		val, exists := store.Get(streamKey)
		if !exists || val.StreamData == nil {
			continue // Skip non-existent streams
		}

		stream := val.StreamData

		// Parse start ID (exclusive - we want entries AFTER this ID)
		startTimestamp, startSequence := parseStreamIDForXRead(startID)

		// Collect entries after the start ID
		var entries [][]byte
		for _, entry := range stream.Entries {
			entryTimestamp, entrySequence := parseEntryID(entry.Id)

			// Check if entry is AFTER start ID (exclusive)
			if isAfterID(entryTimestamp, entrySequence, startTimestamp, startSequence) {
				entries = append(entries, encodeStreamEntry(entry))
			}
		}

		// Only include stream in results if it has entries
		if len(entries) > 0 {
			streamResult := resp.EncodeArray([][]byte{
				resp.EncodeBulkString(streamKey),
				resp.EncodeArray(entries),
			})
			results = append(results, streamResult)
		}
	}
	return results
}

// blockAndWait blocks and waits for new entries in the specified streams
func blockAndWait(keys []string, ids []string, timeoutMs int64) RESPValue {
	// Calculate end time for timeout
	var endTime time.Time
	hasTimeout := timeoutMs > 0
	if hasTimeout {
		endTime = time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	}

	// Poll for new entries
	pollInterval := 100 * time.Millisecond
	if hasTimeout && timeoutMs < 100 {
		pollInterval = time.Duration(timeoutMs) * time.Millisecond
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		// Check for new entries
		results := readStreams(keys, ids)
		if len(results) > 0 {
			return resp.EncodeArray(results)
		}

		// Check timeout
		if hasTimeout && time.Now().After(endTime) {
			return resp.EncodeNullBulkString()
		}

		// Wait for next poll
		<-ticker.C

		// For blocking without timeout (0), continue indefinitely
		// For blocking with timeout, the timeout check above will handle it
	}
}

// parseStreamIDForXRead parses a stream ID for XREAD (similar to parseStreamID but always exclusive)
func parseStreamIDForXRead(id string) (int64, int64) {
	// Handle special case for $
	if id == "$" {
		// $ means the ID of the last entry in the stream
		// This will be handled by the caller
		return math.MaxInt64, math.MaxInt64
	}

	parts := strings.Split(id, "-")
	if len(parts) == 0 {
		return 0, 0
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0
	}

	var sequence int64
	if len(parts) == 1 {
		// No sequence number provided - default to max to get entries with higher timestamps
		sequence = math.MaxInt64
	} else {
		sequence, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			sequence = 0
		}
	}

	return timestamp, sequence
}

// isAfterID checks if an entry comes after the specified ID (exclusive comparison)
func isAfterID(entryTS, entrySeq, afterTS, afterSeq int64) bool {
	if entryTS > afterTS {
		return true
	}
	if entryTS == afterTS && entrySeq > afterSeq {
		return true
	}
	return false
}
