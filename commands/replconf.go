package command

import (
	"fmt"

	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
)

type ReplConfCommand Command

func (cmd *ReplConfCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) < 2 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	switch cmd.args[0] {
	case "listening-port":
		// Store the replica's listening port (used for monitoring/logging)
		// The port value is in cmd.args[1]
		return resp.Success()
	case "capa":
		// Acknowledge capability (e.g., "psync2")
		// The capability value is in cmd.args[1]
		return resp.Success()
	case "getack":
		totalBytes := fmt.Sprint(con.BytesRead)
		return resp.EncodeArrayBulk("replconf", "ACK", totalBytes)
	default:
		return resp.EncodeSimpleError(errSyntax)
	}
}
