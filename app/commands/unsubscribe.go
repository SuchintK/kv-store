package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
)

type UnsubscribeCommand Command

func (cmd *UnsubscribeCommand) Execute(con *client.Client) RESPValue {
	var channel string
	if len(cmd.args) == 0 {
		// Unsubscribe from all channels
		channel = ""
	} else if len(cmd.args) == 1 {
		// Unsubscribe from specific channel
		channel = cmd.args[0]
	} else {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	count := pubsub.Global.Unsubscribe(con, channel)

	return resp.EncodePubSubResponse("unsubscribe", channel, count)
}
