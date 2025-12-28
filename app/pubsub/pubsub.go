package pubsub

import (
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
)

// Manager handles Pub/Sub subscriptions
type Manager struct {
	mu       sync.RWMutex
	channels map[string]map[*client.Client]bool // channel -> set of clients
}

var Global = &Manager{
	channels: make(map[string]map[*client.Client]bool),
}

// Subscribe adds a client to a single channel
func (m *Manager) Subscribe(cli *client.Client, channel string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.channels[channel] == nil {
		m.channels[channel] = make(map[*client.Client]bool)
	}
	m.channels[channel][cli] = true
	cli.Subscribe(channel)
	return cli.SubscriptionCount()
}

// Unsubscribe removes a client from a single channel
// If channel is empty string, unsubscribe from all channels
func (m *Manager) Unsubscribe(cli *client.Client, channel string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If no channel specified, unsubscribe from all
	if channel == "" {
		channels := cli.GetSubscribedChannels()
		for _, ch := range channels {
			if m.channels[ch] != nil {
				delete(m.channels[ch], cli)
				if len(m.channels[ch]) == 0 {
					delete(m.channels, ch)
				}
			}
			cli.Unsubscribe(ch)
		}
	} else {
		// Unsubscribe from specific channel
		if m.channels[channel] != nil {
			delete(m.channels[channel], cli)
			if len(m.channels[channel]) == 0 {
				delete(m.channels, channel)
			}
		}
		cli.Unsubscribe(channel)
	}

	return cli.SubscriptionCount()
}

// Publish sends a message to all clients subscribed to the channel
// Returns the number of clients that received the message
func (m *Manager) Publish(channel, message string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subscribers := m.channels[channel]
	if subscribers == nil {
		return 0
	}

	count := 0
	for cli := range subscribers {
		// Send the message in RESP format
		response := EncodePubSubMessage(channel, message)
		cli.Write(response)
		cli.Flush()
		count++
	}

	return count
}

// UnsubscribeAll removes a client from all channels (used on disconnect)
func (m *Manager) UnsubscribeAll(cli *client.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for channel := range m.channels {
		delete(m.channels[channel], cli)
		if len(m.channels[channel]) == 0 {
			delete(m.channels, channel)
		}
	}
	cli.ClearSubscriptions()
}

// ResetGlobal resets the global pub/sub manager (for testing)
func ResetGlobal() {
	Global = &Manager{
		channels: make(map[string]map[*client.Client]bool),
	}
}

// EncodePubSubMessage creates a RESP array for a pub/sub message
// Format: *3\r\n$7\r\nmessage\r\n$<len>\r\n<channel>\r\n$<len>\r\n<message>\r\n
func EncodePubSubMessage(channel, message string) []byte {
	return resp.EncodeArray([][]byte{
		resp.EncodeBulkString("message"),
		resp.EncodeBulkString(channel),
		resp.EncodeBulkString(message),
	})
}
