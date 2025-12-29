package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type ZRemCommand Command

func (cmd *ZRemCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) < 2 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]
	members := cmd.args[1:]

	// Get sorted set
	val, exists := store.Get(key)
	if !exists || val.SortedSetData == nil {
		return resp.EncodeInteger(0)
	}

	// Remove members
	removed := 0
	for _, member := range members {
		if val.SortedSetData.Remove(member) {
			removed++
		}
	}

	return resp.EncodeInteger(int64(removed))
}
