package command

import (
	"strconv"
	"strings"
	"time"

	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type SetCommand Command

func (cmd *SetCommand) Execute(con *client.Client) RESPValue {
	numArgs := len(cmd.args)
	// TODO write a proper flag parser
	if numArgs != 2 && numArgs != 4 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}
	key := cmd.args[0]
	data := cmd.args[1]
	var expiry int64 = -1

	if numArgs == 4 {
		pxFlag := strings.ToLower(cmd.args[2])
		if pxFlag != "px" {
			return resp.EncodeSimpleError(errSyntax)
		}

		exp, err := strconv.ParseInt(cmd.args[3], 10, 64)
		if err != nil || exp < 0 {
			return resp.EncodeSimpleError("invalid expiry time")
		}
		expiry = exp
	}

	var value *store.Value = &store.Value{}
	value.Data = data
	if expiry > 0 {
		duration := time.Duration(expiry) * time.Millisecond
		expiresAt := time.Now().Add(duration)
		value.ExpiresAt = &expiresAt
	}
	store.Set(key, value)
	return resp.Success()
}
