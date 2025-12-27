package client

import (
	"bufio"
	"net"
)

type Client struct {
	conn net.Conn
	*bufio.Reader
	*bufio.Writer
	BytesRead int
	// Transaction state
	InTransaction  bool
	QueuedCommands []QueuedCommand
}

type QueuedCommand struct {
	Label string
	Args  []string
}

func New(conn net.Conn) Client {
	return Client{
		conn:           conn,
		Reader:         bufio.NewReader(conn),
		Writer:         bufio.NewWriter(conn),
		BytesRead:      0,
		InTransaction:  false,
		QueuedCommands: make([]QueuedCommand, 0),
	}
}

func (c *Client) Connection() net.Conn {
	return c.conn
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) StartTransaction() {
	c.InTransaction = true
	c.QueuedCommands = make([]QueuedCommand, 0)
}

func (c *Client) QueueCommand(label string, args []string) {
	c.QueuedCommands = append(c.QueuedCommands, QueuedCommand{
		Label: label,
		Args:  args,
	})
}

func (c *Client) DiscardTransaction() {
	c.InTransaction = false
	c.QueuedCommands = make([]QueuedCommand, 0)
}

func (c *Client) GetQueuedCommands() []QueuedCommand {
	return c.QueuedCommands
}

func (c *Client) IsInTransaction() bool {
	return c.InTransaction
}
