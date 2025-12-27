package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type XAddCommand Command

func (cmd *XAddCommand) Execute(con *client.Client) RESPValue {
	numArgs := len(cmd.args)
	// TODO write a proper flag parser
	if numArgs < 4 || (numArgs-2)%2 != 0 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]
	idArg := cmd.args[1]

	// Parse field-value pairs
	fields := make(map[string]string)
	for i := 2; i < numArgs; i += 2 {
		fields[cmd.args[i]] = cmd.args[i+1]
	}

	// Get or create stream
	var val *store.Value
	var stream *store.Stream
	val, exists := store.Get(key)

	if exists {
		stream = val.StreamData
	} else {
		// Create new stream
		stream = &store.Stream{
			Entries:       make([]*store.StreamEntry, 0),
			LastTimestamp: 0,
			LastSequence:  0,
		}
		val = &store.Value{
			StreamData: stream,
		}
		store.Set(key, val)
	}

	// Generate or validate ID
	var entryID string
	if idArg == "*" {
		// Auto-generate ID using current timestamp
		timestamp := time.Now().UnixMilli()
		sequence := int64(0)

		if timestamp == stream.LastTimestamp {
			sequence = stream.LastSequence + 1
		} else if timestamp < stream.LastTimestamp {
			// Use last timestamp + 1
			timestamp = stream.LastTimestamp
			sequence = stream.LastSequence + 1
		}

		stream.LastTimestamp = timestamp
		stream.LastSequence = sequence
		entryID = fmt.Sprintf("%d-%d", timestamp, sequence)
	} else {
		// Use provided ID
		// Basic validation - should be in format timestamp-sequence
		parts := strings.Split(idArg, "-")
		if len(parts) != 2 {
			return resp.EncodeSimpleError(invalidStreamID)
		}

		timestamp, err1 := strconv.ParseInt(parts[0], 10, 64)
		sequence, err2 := strconv.ParseInt(parts[1], 10, 64)

		if err1 != nil || err2 != nil {
			return resp.EncodeSimpleError(invalidStreamID)
		}

		// Check if ID is greater than last ID
		if len(stream.Entries) > 0 {
			if timestamp < stream.LastTimestamp || (timestamp == stream.LastTimestamp && sequence <= stream.LastSequence) {
				return resp.EncodeSimpleError(idGreaterThanTopElement)
			}
		}

		stream.LastTimestamp = timestamp
		stream.LastSequence = sequence
		entryID = idArg
	}

	// Create and add entry
	entry := &store.StreamEntry{
		Id:     entryID,
		Fields: fields,
	}
	stream.Entries = append(stream.Entries, entry)
	return resp.EncodeBulkString(entryID)
}
