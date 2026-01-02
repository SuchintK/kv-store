package command

import (
	"strconv"

	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type ZRangeCommand Command

func (cmd *ZRangeCommand) Execute(con *client.Client) RESPValue {
	numArgs := len(cmd.args)

	// ZRANGE key start stop [WITHSCORES]
	if numArgs < 3 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]
	startStr := cmd.args[1]
	stopStr := cmd.args[2]
	withScores := false

	if numArgs >= 4 && cmd.args[3] == "withscores" {
		withScores = true
	}

	// Parse start and stop
	start, err := strconv.Atoi(startStr)
	if err != nil {
		return resp.EncodeSimpleError("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(stopStr)
	if err != nil {
		return resp.EncodeSimpleError("ERR value is not an integer or out of range")
	}

	// Get sorted set
	val, exists := store.Get(key)
	if !exists || val.SortedSetData == nil {
		return resp.EncodeArray([][]byte{})
	}

	zset := val.SortedSetData

	if withScores {
		// Return with scores
		rangeWithScores := zset.GetRangeWithScores(start, stop)
		result := make([][]byte, 0, len(rangeWithScores)*2)
		for _, item := range rangeWithScores {
			result = append(result, resp.EncodeBulkString(item.Member))
			result = append(result, resp.EncodeBulkString(strconv.FormatFloat(item.Score, 'f', -1, 64)))
		}
		return resp.EncodeArray(result)
	}

	// Return only members
	members := zset.GetRange(start, stop)
	result := make([][]byte, len(members))
	for i, member := range members {
		result[i] = resp.EncodeBulkString(member)
	}
	return resp.EncodeArray(result)
}
