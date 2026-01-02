package command

import (
	"strconv"

	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type LRangeCommand Command

func (cmd *LRangeCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) != 3 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]
	startStr := cmd.args[1]
	stopStr := cmd.args[2]

	// Parse start and stop
	start, err := strconv.Atoi(startStr)
	if err != nil {
		return resp.EncodeSimpleError("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(stopStr)
	if err != nil {
		return resp.EncodeSimpleError("ERR value is not an integer or out of range")
	}

	val, exists := store.Get(key)

	// Key doesn't exist
	if !exists {
		return resp.EncodeArray([][]byte{})
	}

	// Check if it's not a list
	if val.ListData == nil {
		return resp.EncodeSimpleError(errWrongType)
	}

	length := len(val.ListData)

	// Handle negative indices
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = length + stop
	}

	// Handle out of bounds
	if start > stop || start >= length {
		return resp.EncodeArray([][]byte{})
	}
	if stop >= length {
		stop = length - 1
	}

	// Get range
	result := make([][]byte, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, resp.EncodeBulkString(val.ListData[i]))
	}

	return resp.EncodeArray(result)
}
