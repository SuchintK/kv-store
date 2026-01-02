package command

import (
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
)

type DiscardCommand Command

func (cmd *DiscardCommand) Execute(con *client.Client) RESPValue {
	if !con.IsInTransaction() {
		return resp.EncodeSimpleError("ERR DISCARD without MULTI")
	}

	con.DiscardTransaction()
	return resp.EncodeSimpleString("OK")
}
