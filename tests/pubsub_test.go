package tests

import (
	"bytes"
	"net"
	"testing"
	"time"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/pubsub"
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	*bytes.Buffer
}

func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func newMockClient() *client.Client {
	conn := &mockConn{Buffer: &bytes.Buffer{}}
	cli := client.New(conn)
	return &cli
}

func TestSubscribeCommand(t *testing.T) {
	// Reset global pubsub manager
	pubsub.ResetGlobal()

	cli := newMockClient()

	// Test SUBSCRIBE with valid channel
	cmd := command.New("subscribe", []string{"test-channel"})

	response := cmd.Execute(cli)

	// Should return subscription confirmation
	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	// Check if client is subscribed
	if !cli.IsSubscribed() {
		t.Error("Client should be in subscribed mode")
	}

	if cli.SubscriptionCount() != 1 {
		t.Errorf("Expected 1 subscription, got %d", cli.SubscriptionCount())
	}
}

func TestPublishCommand(t *testing.T) {
	// Reset global pubsub manager
	pubsub.ResetGlobal()

	// Create subscriber client
	subscriber := newMockClient()

	// Subscribe to channel
	subCmd := command.New("subscribe", []string{"test-channel"})
	subCmd.Execute(subscriber)

	// Create publisher client
	publisher := newMockClient()

	// Test PUBLISH command
	pubCmd := command.New("publish", []string{"test-channel", "Hello, World!"})

	response := pubCmd.Execute(publisher)

	// Should return number of subscribers (1)
	expected := resp.EncodeInteger(1)
	if !bytes.Equal(response, expected) {
		t.Errorf("Expected %s, got %s", string(expected), string(response))
	}
}

func TestUnsubscribeCommand(t *testing.T) {
	// Reset global pubsub manager
	pubsub.ResetGlobal()

	cli := newMockClient()

	// First subscribe to a channel
	subCmd := command.New("subscribe", []string{"test-channel"})
	subCmd.Execute(cli)

	// Verify subscription
	if cli.SubscriptionCount() != 1 {
		t.Fatalf("Expected 1 subscription, got %d", cli.SubscriptionCount())
	}

	// Test UNSUBSCRIBE from specific channel
	unsubCmd := command.New("unsubscribe", []string{"test-channel"})

	response := unsubCmd.Execute(cli)

	// Should return unsubscription confirmation
	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	// Check if client is no longer subscribed
	if cli.IsSubscribed() {
		t.Error("Client should not be in subscribed mode")
	}

	if cli.SubscriptionCount() != 0 {
		t.Errorf("Expected 0 subscriptions, got %d", cli.SubscriptionCount())
	}
}

func TestMultipleSubscribersToSameChannel(t *testing.T) {
	// Reset global pubsub manager
	pubsub.ResetGlobal()

	// Create three subscribers
	subscriber1 := newMockClient()
	subscriber2 := newMockClient()
	subscriber3 := newMockClient()

	// All subscribe to the same channel
	subCmd1 := command.New("subscribe", []string{"news"})
	subCmd1.Execute(subscriber1)

	subCmd2 := command.New("subscribe", []string{"news"})
	subCmd2.Execute(subscriber2)

	subCmd3 := command.New("subscribe", []string{"news"})
	subCmd3.Execute(subscriber3)

	// Verify all are subscribed
	if !subscriber1.IsSubscribed() || !subscriber2.IsSubscribed() || !subscriber3.IsSubscribed() {
		t.Error("All subscribers should be in subscribed mode")
	}

	// Create publisher and publish to the channel
	publisher := newMockClient()
	pubCmd := command.New("publish", []string{"news", "Breaking news!"})
	response := pubCmd.Execute(publisher)

	// Should return number of subscribers (3)
	expected := resp.EncodeInteger(3)
	if !bytes.Equal(response, expected) {
		t.Errorf("Expected %s, got %s", string(expected), string(response))
	}
}

func TestSubscribeToMultipleChannels(t *testing.T) {
	// Reset global pubsub manager
	pubsub.ResetGlobal()

	cli := newMockClient()

	// Subscribe to first channel
	subCmd1 := command.New("subscribe", []string{"channel1"})
	subCmd1.Execute(cli)

	// Subscribe to second channel
	subCmd2 := command.New("subscribe", []string{"channel2"})
	subCmd2.Execute(cli)

	// Subscribe to third channel
	subCmd3 := command.New("subscribe", []string{"channel3"})
	subCmd3.Execute(cli)

	// Verify subscription count
	if cli.SubscriptionCount() != 3 {
		t.Errorf("Expected 3 subscriptions, got %d", cli.SubscriptionCount())
	}

	// Verify client is in subscribed mode
	if !cli.IsSubscribed() {
		t.Error("Client should be in subscribed mode")
	}
}

func TestPublishToMultipleChannels(t *testing.T) {
	// Reset global pubsub manager
	pubsub.ResetGlobal()

	// Create subscribers for different channels
	subscriber1 := newMockClient()
	subscriber2 := newMockClient()
	subscriber3 := newMockClient()

	// Subscribe to different channels
	command.New("subscribe", []string{"sports"}).Execute(subscriber1)
	command.New("subscribe", []string{"tech"}).Execute(subscriber2)
	command.New("subscribe", []string{"weather"}).Execute(subscriber3)

	// Create publisher
	publisher := newMockClient()

	// Publish to sports channel
	response1 := command.New("publish", []string{"sports", "Goal!"}).Execute(publisher)
	expected1 := resp.EncodeInteger(1)
	if !bytes.Equal(response1, expected1) {
		t.Errorf("Sports channel: Expected %s, got %s", string(expected1), string(response1))
	}

	// Publish to tech channel
	response2 := command.New("publish", []string{"tech", "New release"}).Execute(publisher)
	expected2 := resp.EncodeInteger(1)
	if !bytes.Equal(response2, expected2) {
		t.Errorf("Tech channel: Expected %s, got %s", string(expected2), string(response2))
	}

	// Publish to weather channel
	response3 := command.New("publish", []string{"weather", "Sunny"}).Execute(publisher)
	expected3 := resp.EncodeInteger(1)
	if !bytes.Equal(response3, expected3) {
		t.Errorf("Weather channel: Expected %s, got %s", string(expected3), string(response3))
	}
}

func TestPublishToChannelWithNoSubscribers(t *testing.T) {
	// Reset global pubsub manager
	pubsub.ResetGlobal()

	publisher := newMockClient()

	// Publish to channel with no subscribers
	pubCmd := command.New("publish", []string{"empty-channel", "Hello?"})
	response := pubCmd.Execute(publisher)

	// Should return 0 subscribers
	expected := resp.EncodeInteger(0)
	if !bytes.Equal(response, expected) {
		t.Errorf("Expected %s, got %s", string(expected), string(response))
	}
}

func TestUnsubscribeFromOneOfMultipleChannels(t *testing.T) {
	// Reset global pubsub manager
	pubsub.ResetGlobal()

	cli := newMockClient()

	// Subscribe to three channels
	command.New("subscribe", []string{"channel1"}).Execute(cli)
	command.New("subscribe", []string{"channel2"}).Execute(cli)
	command.New("subscribe", []string{"channel3"}).Execute(cli)

	// Verify 3 subscriptions
	if cli.SubscriptionCount() != 3 {
		t.Fatalf("Expected 3 subscriptions, got %d", cli.SubscriptionCount())
	}

	// Unsubscribe from one channel
	unsubCmd := command.New("unsubscribe", []string{"channel2"})
	unsubCmd.Execute(cli)

	// Should still have 2 subscriptions
	if cli.SubscriptionCount() != 2 {
		t.Errorf("Expected 2 subscriptions after unsubscribe, got %d", cli.SubscriptionCount())
	}

	// Should still be in subscribed mode
	if !cli.IsSubscribed() {
		t.Error("Client should still be in subscribed mode")
	}

	// Unsubscribe from remaining channels
	command.New("unsubscribe", []string{"channel1"}).Execute(cli)
	command.New("unsubscribe", []string{"channel3"}).Execute(cli)

	// Should have 0 subscriptions
	if cli.SubscriptionCount() != 0 {
		t.Errorf("Expected 0 subscriptions, got %d", cli.SubscriptionCount())
	}

	// Should not be in subscribed mode
	if cli.IsSubscribed() {
		t.Error("Client should not be in subscribed mode")
	}
}

func TestMultipleSubscribersDifferentChannels(t *testing.T) {
	// Reset global pubsub manager
	pubsub.ResetGlobal()

	// Create subscribers
	subscriber1 := newMockClient()
	subscriber2 := newMockClient()
	subscriber3 := newMockClient()

	// subscriber1 subscribes to channel1 and channel2
	command.New("subscribe", []string{"channel1"}).Execute(subscriber1)
	command.New("subscribe", []string{"channel2"}).Execute(subscriber1)

	// subscriber2 subscribes to channel2 and channel3
	command.New("subscribe", []string{"channel2"}).Execute(subscriber2)
	command.New("subscribe", []string{"channel3"}).Execute(subscriber2)

	// subscriber3 subscribes to channel1 and channel3
	command.New("subscribe", []string{"channel1"}).Execute(subscriber3)
	command.New("subscribe", []string{"channel3"}).Execute(subscriber3)

	// Verify subscription counts
	if subscriber1.SubscriptionCount() != 2 {
		t.Errorf("Subscriber1: Expected 2 subscriptions, got %d", subscriber1.SubscriptionCount())
	}
	if subscriber2.SubscriptionCount() != 2 {
		t.Errorf("Subscriber2: Expected 2 subscriptions, got %d", subscriber2.SubscriptionCount())
	}
	if subscriber3.SubscriptionCount() != 2 {
		t.Errorf("Subscriber3: Expected 2 subscriptions, got %d", subscriber3.SubscriptionCount())
	}

	// Publish to channel1 (should reach subscriber1 and subscriber3)
	publisher := newMockClient()
	response1 := command.New("publish", []string{"channel1", "Message 1"}).Execute(publisher)
	expected1 := resp.EncodeInteger(2)
	if !bytes.Equal(response1, expected1) {
		t.Errorf("Channel1: Expected %s, got %s", string(expected1), string(response1))
	}

	// Publish to channel2 (should reach subscriber1 and subscriber2)
	response2 := command.New("publish", []string{"channel2", "Message 2"}).Execute(publisher)
	expected2 := resp.EncodeInteger(2)
	if !bytes.Equal(response2, expected2) {
		t.Errorf("Channel2: Expected %s, got %s", string(expected2), string(response2))
	}

	// Publish to channel3 (should reach subscriber2 and subscriber3)
	response3 := command.New("publish", []string{"channel3", "Message 3"}).Execute(publisher)
	expected3 := resp.EncodeInteger(2)
	if !bytes.Equal(response3, expected3) {
		t.Errorf("Channel3: Expected %s, got %s", string(expected3), string(response3))
	}
}

func TestUnsubscribeFromAllChannels(t *testing.T) {
	// Reset global pubsub manager
	pubsub.ResetGlobal()

	cli := newMockClient()

	// Subscribe to multiple channels
	command.New("subscribe", []string{"channel1"}).Execute(cli)
	command.New("subscribe", []string{"channel2"}).Execute(cli)
	command.New("subscribe", []string{"channel3"}).Execute(cli)

	// Verify 3 subscriptions
	if cli.SubscriptionCount() != 3 {
		t.Fatalf("Expected 3 subscriptions, got %d", cli.SubscriptionCount())
	}

	// Unsubscribe from all channels (empty args)
	unsubCmd := command.New("unsubscribe", []string{})
	unsubCmd.Execute(cli)

	// Should have 0 subscriptions
	if cli.SubscriptionCount() != 0 {
		t.Errorf("Expected 0 subscriptions after unsubscribe all, got %d", cli.SubscriptionCount())
	}

	// Should not be in subscribed mode
	if cli.IsSubscribed() {
		t.Error("Client should not be in subscribed mode after unsubscribing from all")
	}
}
