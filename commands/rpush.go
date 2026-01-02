package command

import (
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type RPushCommand Command

func (cmd *RPushCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) < 2 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]
	elements := cmd.args[1:]

	val, exists := store.Get(key)

	// Check if key exists and is not a list
	if exists && val.ListData == nil && val.Data != "" {
		return resp.EncodeSimpleError(errWrongType)
	}

	// Initialize list if it doesn't exist
	if !exists || val.ListData == nil {
		val = &store.Value{
			ListData: []string{},
		}
	}

	// Push elements to the tail (append)
	val.ListData = append(val.ListData, elements...)

	store.Set(key, val)

	return resp.EncodeInteger(int64(len(val.ListData)))
}
