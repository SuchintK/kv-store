package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
)

type PublishCommand Command

func (cmd *PublishCommand) Execute(con *client.Client) RESPValue {
	if len(cmd.args) < 2 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	channel := cmd.args[0]
	message := cmd.args[1]

	// Publish the message and get the number of subscribers who received it
	count := pubsub.Global.Publish(channel, message)

	return resp.EncodeInteger(int64(count))
}
