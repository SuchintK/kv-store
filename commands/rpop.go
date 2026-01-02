package command

import (
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type RPopCommand Command

func (cmd *RPopCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) != 1 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]
	val, exists := store.Get(key)

	// Key doesn't exist
	if !exists || val.ListData == nil {
		return resp.EncodeNullBulkString()
	}

	// Check if it's not a list
	if val.ListData == nil {
		return resp.EncodeSimpleError(errWrongType)
	}

	// Empty list
	if len(val.ListData) == 0 {
		return resp.EncodeNullBulkString()
	}

	// Pop the last element
	lastIndex := len(val.ListData) - 1
	element := val.ListData[lastIndex]
	val.ListData = val.ListData[:lastIndex]

	// Update or delete the key
	if len(val.ListData) == 0 {
		store.Delete(key)
	} else {
		store.Set(key, val)
	}

	return resp.EncodeBulkString(element)
}
