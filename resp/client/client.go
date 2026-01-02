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
	// Pub/Sub state
	subscribedChannels map[string]bool
}

type QueuedCommand struct {
	Label string
	Args  []string
}

func New(conn net.Conn) Client {
	return Client{
		conn:               conn,
		Reader:             bufio.NewReader(conn),
		Writer:             bufio.NewWriter(conn),
		BytesRead:          0,
		InTransaction:      false,
		QueuedCommands:     make([]QueuedCommand, 0),
		subscribedChannels: make(map[string]bool),
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

// Pub/Sub methods

func (c *Client) Subscribe(channel string) {
	c.subscribedChannels[channel] = true
}

func (c *Client) Unsubscribe(channel string) {
	delete(c.subscribedChannels, channel)
}

func (c *Client) IsSubscribed() bool {
	return len(c.subscribedChannels) > 0
}

func (c *Client) SubscriptionCount() int {
	return len(c.subscribedChannels)
}

func (c *Client) GetSubscribedChannels() []string {
	channels := make([]string, 0, len(c.subscribedChannels))
	for ch := range c.subscribedChannels {
		channels = append(channels, ch)
	}
	return channels
}

func (c *Client) ClearSubscriptions() {
	c.subscribedChannels = make(map[string]bool)
}
