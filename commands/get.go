package command

import (
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type GetCommand Command

func (cmd *GetCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) != 1 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}
	item, exist := store.Get(cmd.args[0])
	if !exist {
		return resp.EncodeNullBulkString()
	}
	return resp.EncodeBulkString(item.Data)
}
