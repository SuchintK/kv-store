package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
)

type PingCommand Command

func (cmd *PingCommand) Execute(con *client.Client) RESPValue {
	// In subscribed mode, PING returns a different format
	if con.IsSubscribed() {
		if len(cmd.args) == 0 {
			// Returns: *2\r\n$4\r\npong\r\n$0\r\n\r\n
			return resp.EncodeArray([][]byte{
				resp.EncodeBulkString("pong"),
				resp.EncodeBulkString(""),
			})
		}
		// With message: *2\r\n$4\r\npong\r\n$<len>\r\n<message>\r\n
		return resp.EncodeArray([][]byte{
			resp.EncodeBulkString("pong"),
			resp.EncodeBulkString(cmd.args[0]),
		})
	}

	if len(cmd.args) == 0 {
		return resp.EncodeSimpleString("PONG")
	}
	if len(cmd.args) == 1 {
		return resp.EncodeBulkString(cmd.args[0])
	}
	return resp.EncodeSimpleError(errWrongNumberOfArgs)
}
