package command

import (
	"math"
	"strconv"
	"strings"

	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type XRangeCommand Command

func (cmd *XRangeCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) != 3 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]
	startID := cmd.args[1]
	endID := cmd.args[2]

	// Get stream
	val, exists := store.Get(key)
	if !exists || val.StreamData == nil {
		return resp.EncodeArray([][]byte{})
	}

	stream := val.StreamData

	// Parse start and end IDs
	startTimestamp, startSequence := parseStreamID(startID, true)
	endTimestamp, endSequence := parseStreamID(endID, false)

	// Filter entries within range
	var results [][]byte
	for _, entry := range stream.Entries {
		entryTimestamp, entrySequence := parseEntryID(entry.Id)

		// Check if entry is within range
		if isInRange(entryTimestamp, entrySequence, startTimestamp, startSequence, endTimestamp, endSequence) {
			results = append(results, encodeStreamEntry(entry))
		}
	}

	return resp.EncodeArray(results)
}

// parseStreamID parses a stream ID string and returns timestamp and sequence
// isStart indicates if this is the start ID (for default sequence handling)
func parseStreamID(id string, isStart bool) (int64, int64) {
	// Handle special cases
	if id == "-" {
		return 0, 0
	}
	if id == "+" {
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
		// No sequence number provided
		if isStart {
			sequence = 0 // Start defaults to 0
		} else {
			sequence = math.MaxInt64 // End defaults to max
		}
	} else {
		sequence, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			sequence = 0
		}
	}

	return timestamp, sequence
}

// parseEntryID parses an entry's ID into timestamp and sequence
func parseEntryID(id string) (int64, int64) {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0
	}

	timestamp, err1 := strconv.ParseInt(parts[0], 10, 64)
	sequence, err2 := strconv.ParseInt(parts[1], 10, 64)

	if err1 != nil || err2 != nil {
		return 0, 0
	}

	return timestamp, sequence
}

// isInRange checks if an entry is within the specified range
func isInRange(entryTS, entrySeq, startTS, startSeq, endTS, endSeq int64) bool {
	// Check if entry >= start
	if entryTS < startTS {
		return false
	}
	if entryTS == startTS && entrySeq < startSeq {
		return false
	}

	// Check if entry <= end
	if entryTS > endTS {
		return false
	}
	if entryTS == endTS && entrySeq > endSeq {
		return false
	}

	return true
}

// encodeStreamEntry encodes a stream entry as a RESP array
func encodeStreamEntry(entry *store.StreamEntry) []byte {
	// Entry is encoded as [ID, [field1, value1, field2, value2, ...]]
	fields := make([][]byte, 0, len(entry.Fields)*2)
	for key, val := range entry.Fields {
		fields = append(fields, resp.EncodeBulkString(key))
		fields = append(fields, resp.EncodeBulkString(val))
	}

	return resp.EncodeArray([][]byte{
		resp.EncodeBulkString(entry.Id),
		resp.EncodeArray(fields),
	})
}
