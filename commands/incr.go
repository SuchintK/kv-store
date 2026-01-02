package command

import (
	"strconv"

	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type IncrCommand Command

func (cmd *IncrCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) != 1 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]

	// Get current value
	item, exists := store.Get(key)
	var currentValue int64 = 0

	if exists && item.Data != "" {
		// Try to parse existing value as integer
		val, err := strconv.ParseInt(item.Data, 10, 64)
		if err != nil {
			return resp.EncodeSimpleError("ERR value is not an integer or out of range")
		}
		currentValue = val
	}

	// Increment
	currentValue++

	// Store new value
	newValue := &store.Value{
		Data: strconv.FormatInt(currentValue, 10),
	}
	store.Set(key, newValue)

	// Return new value as integer
	return resp.EncodeInteger(currentValue)
}
