package command

import (
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type LPushCommand Command

func (cmd *LPushCommand) Execute(con *client.Client) RESPValue {
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

	// Push elements to the head (prepend in reverse order to maintain order)
	newList := make([]string, 0, len(val.ListData)+len(elements))
	for i := len(elements) - 1; i >= 0; i-- {
		newList = append(newList, elements[i])
	}
	newList = append(newList, val.ListData...)
	val.ListData = newList

	store.Set(key, val)

	return resp.EncodeInteger(int64(len(val.ListData)))
}
