package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
)

type ExecCommand Command

func (cmd *ExecCommand) Execute(con *client.Client) RESPValue {
	if !con.IsInTransaction() {
		return resp.EncodeSimpleError("ERR EXEC without MULTI")
	}

	queuedCommands := con.GetQueuedCommands()
	con.DiscardTransaction()

	// Execute all queued commands
	results := make([][]byte, 0, len(queuedCommands))

	for _, qCmd := range queuedCommands {
		executor := New(qCmd.Label, qCmd.Args)

		// Execute command and capture result (including errors)
		// In Redis, errors within transactions don't stop execution
		result := executor.Execute(con)
		results = append(results, result)
	}

	return resp.EncodeArray(results)
}
