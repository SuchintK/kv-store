package command

import (
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
)

const (
	errWrongNumberOfArgs    = "wrong number of arguments"
	errSyntax               = "syntax error"
	errWrongType            = "WRONGTYPE Operation against a key holding the wrong kind of value"
	invalidStreamID         = "ERR Invalid stream ID specified as stream command argument"
	idGreaterThanTopElement = "ERR The ID specified in XADD is equal or smaller than the target stream top item"
	errSubscribedMode       = "ERR only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context"
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

// IsAllowedInSubscribedMode checks if a command is allowed when client is in subscribed mode
func IsAllowedInSubscribedMode(label string) bool {
	switch label {
	case "subscribe", "unsubscribe", "psubscribe", "punsubscribe", "ping", "quit", "reset":
		return true
	default:
		return false
	}
}

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
	case "subscribe":
		return &SubscribeCommand{label: label, args: params}
	case "unsubscribe":
		return &UnsubscribeCommand{label: label, args: params}
	case "publish":
		return &PublishCommand{label: label, args: params}
	case "zadd":
		return &ZAddCommand{label: label, args: params, IsMutation: true}
	case "zrank":
		return &ZRankCommand{label: label, args: params}
	case "zrange":
		return &ZRangeCommand{label: label, args: params}
	case "zcard":
		return &ZCardCommand{label: label, args: params}
	case "zscore":
		return &ZScoreCommand{label: label, args: params}
	case "zrem":
		return &ZRemCommand{label: label, args: params, IsMutation: true}
	case "lpush":
		return &LPushCommand{label: label, args: params, IsMutation: true}
	case "rpush":
		return &RPushCommand{label: label, args: params, IsMutation: true}
	case "lpop":
		return &LPopCommand{label: label, args: params, IsMutation: true}
	case "rpop":
		return &RPopCommand{label: label, args: params, IsMutation: true}
	case "llen":
		return &LLenCommand{label: label, args: params}
	case "lrange":
		return &LRangeCommand{label: label, args: params}
	case "blpop":
		return &BLPopCommand{label: label, args: params, IsMutation: true}
	case "geoadd":
		return &GeoAddCommand{label: label, args: params, IsMutation: true}
	case "geopos":
		return &GeoPosCommand{label: label, args: params}
	case "geodist":
		return &GeoDistCommand{label: label, args: params}
	case "georadius":
		return &GeoRadiusCommand{label: label, args: params}
	}
	return &NotImplementedCommand{}
}

func (cmd *NotImplementedCommand) Execute(con *client.Client) RESPValue {
	return resp.EncodeSimpleError("unknown command, may not be implemented yet")
}
