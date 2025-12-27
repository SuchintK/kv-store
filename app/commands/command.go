package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
)

const (
	errWrongNumberOfArgs    = "wrong number of arguments"
	errSyntax               = "syntax error"
	invalidStreamID         = "ERR Invalid stream ID specified as stream command argument"
	idGreaterThanTopElement = "ERR The ID specified in XADD is equal or smaller than the target stream top item"
)

type RESPValue []byte

type Executor interface {
	Execute(client *client.Client) RESPValue
}

type Command struct {
	label      string
	args       []string
	IsMutation bool
}

type NotImplementedCommand Command

func New(label string, params []string) Executor {
	switch label {
	case "ping":
		return &PingCommand{label: label, args: params}
	case "echo":
		return &EchoCommand{label: label, args: params}
	case "set":
		return &SetCommand{label: label, args: params, IsMutation: true}
	case "get":
		return &GetCommand{label: label, args: params}
	case "info":
		return &InfoCommand{label: label, args: params}
	case "replconf":
		return &ReplConfCommand{label: label, args: params}
	case "psync":
		return &PSYNCCommand{label: label, args: params}
	case "xadd":
		return &XAddCommand{label: label, args: params}
	case "xrange":
		return &XRangeCommand{label: label, args: params}
	case "xread":
		return &XReadCommand{label: label, args: params}
	case "incr":
		return &IncrCommand{label: label, args: params, IsMutation: true}
	case "multi":
		return &MultiCommand{label: label, args: params}
	case "exec":
		return &ExecCommand{label: label, args: params}
	case "discard":
		return &DiscardCommand{label: label, args: params}
	}
	return &NotImplementedCommand{}
}

func (cmd *NotImplementedCommand) Execute(con *client.Client) RESPValue {
	return resp.EncodeSimpleError("unknown command, may not be implemented yet")
}
