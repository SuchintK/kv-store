package command

import (
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type InfoCommand Command

func (cmd *InfoCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) > 1 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	if len(cmd.args) != 0 && cmd.args[0] != "replication" {
		return resp.EncodeSimpleError(errSyntax)
	}

	return resp.EncodeBulkString(store.Info.String())
}
