package command

import (
	"strconv"

	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type ZAddCommand Command

func (cmd *ZAddCommand) Execute(con *client.Client) RESPValue {
	numArgs := len(cmd.args)

	// ZADD key score member [score member ...]
	if numArgs < 3 || (numArgs-1)%2 != 0 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]

	// Get or create sorted set
	val, exists := store.Get(key)
	var zset *store.SortedSet

	if exists {
		if val.SortedSetData != nil {
			zset = val.SortedSetData
		} else if val.Data != "" || val.StreamData != nil {
			// Key exists but is not a sorted set
			return resp.EncodeSimpleError("WRONGTYPE Operation against a key holding the wrong kind of value")
		} else {
			// Empty value, create new sorted set
			zset = store.NewSortedSet()
			val.SortedSetData = zset
			store.Set(key, val)
		}
	} else {
		// Create new sorted set
		zset = store.NewSortedSet()
		store.Set(key, &store.Value{
			SortedSetData: zset,
		})
	}

	// Parse and add score-member pairs
	added := 0
	for i := 1; i < numArgs; i += 2 {
		scoreStr := cmd.args[i]
		member := cmd.args[i+1]

		// Parse score
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			return resp.EncodeSimpleError("ERR value is not a valid float")
		}

		// Add to sorted set
		if zset.Add(score, member) {
			added++
		}
	}

	return resp.EncodeInteger(int64(added))
}
