package command

import (
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
)

type MultiCommand Command

func (cmd *MultiCommand) Execute(con *client.Client) RESPValue {
	if con.IsInTransaction() {
		return resp.EncodeSimpleError("ERR MULTI calls can not be nested")
	}

	con.StartTransaction()
	return resp.EncodeSimpleString("OK")
}
