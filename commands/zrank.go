package command

import (
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type ZRankCommand Command

func (cmd *ZRankCommand) Execute(con *client.Client) RESPValue {
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

	rank := val.SortedSetData.GetRank(member)
	if rank == -1 {
		return resp.EncodeNullBulkString()
	}

	return resp.EncodeInteger(int64(rank))
}
