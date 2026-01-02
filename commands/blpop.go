package command

import (
	"strconv"
	"time"

	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type BLPopCommand Command

func (cmd *BLPopCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) < 2 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	// Last argument is timeout
	timeoutStr := cmd.args[len(cmd.args)-1]
	timeout, err := strconv.ParseFloat(timeoutStr, 64)
	if err != nil {
		return resp.EncodeSimpleError("ERR timeout is not a float or out of range")
	}

	// Keys are all arguments except the last one
	keys := cmd.args[:len(cmd.args)-1]

	// Calculate deadline
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(time.Duration(timeout * float64(time.Second)))
	} else {
		// timeout = 0 means block indefinitely (set a very long timeout)
		deadline = time.Now().Add(24 * time.Hour)
	}

	// Poll for elements with a small sleep interval
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		// Try to pop from each key in order
		for _, key := range keys {
			val, exists := store.Get(key)

			if exists && val.ListData != nil && len(val.ListData) > 0 {
				// Pop the first element
				element := val.ListData[0]
				val.ListData = val.ListData[1:]

				// Update or delete the key
				if len(val.ListData) == 0 {
					store.Delete(key)
				} else {
					store.Set(key, val)
				}

				// Return [key, element]
				result := [][]byte{
					resp.EncodeBulkString(key),
					resp.EncodeBulkString(element),
				}
				return resp.EncodeArray(result)
			}
		}

		// Check timeout
		if time.Now().After(deadline) {
			return resp.EncodeNullBulkString()
		}

		// Wait before next poll
		<-ticker.C
	}
}
