package command

import (
	"strconv"

	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type ZScoreCommand Command

func (cmd *ZScoreCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) != 2 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]
	member := cmd.args[1]

	// Get sorted set
	val, exists := store.Get(key)
	if !exists || val.SortedSetData == nil {
		return resp.EncodeNullBulkString()
	}

	score, exists := val.SortedSetData.GetScore(member)
	if !exists {
		return resp.EncodeNullBulkString()
	}

	return resp.EncodeBulkString(strconv.FormatFloat(score, 'f', -1, 64))
}
