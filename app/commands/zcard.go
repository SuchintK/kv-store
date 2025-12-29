package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type ZCardCommand Command

func (cmd *ZCardCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) != 1 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]

	// Get sorted set
	val, exists := store.Get(key)
	if !exists || val.SortedSetData == nil {
		return resp.EncodeInteger(0)
	}

	return resp.EncodeInteger(int64(val.SortedSetData.Card()))
}
