package tests

import (
	"strings"
	"testing"
	"time"

	command "github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func TestXReadCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
		validate func(*testing.T, string)
		timeout  time.Duration
	}{
		{
			name: "Read from single stream",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "1000-0", Fields: map[string]string{"key": "value1"}},
						{Id: "2000-0", Fields: map[string]string{"key": "value2"}},
					},
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args: []string{"streams", "mystream", "0-0"},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "mystream") {
					t.Error("Should contain stream name")
				}
				if !strings.Contains(result, "1000-0") || !strings.Contains(result, "2000-0") {
					t.Error("Should contain entries after 0-0")
				}
			},
		},
		{
			name: "Read from multiple streams",
			setup: func() {
				stream1 := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "1000-0", Fields: map[string]string{"key": "value1"}},
					},
				}
				stream2 := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "2000-0", Fields: map[string]string{"key": "value2"}},
					},
				}
				store.Set("stream1", &store.Value{StreamData: stream1})
				store.Set("stream2", &store.Value{StreamData: stream2})
			},
			args: []string{"streams", "stream1", "stream2", "0-0", "0-0"},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "stream1") || !strings.Contains(result, "stream2") {
					t.Error("Should contain both stream names")
				}
			},
		},
		{
			name: "Blocking read with timeout - receives data",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{},
					LastID:  "0-0",
				}
				store.Set("mystream", &store.Value{StreamData: stream})
				// Add entry after 100ms
				go func() {
					time.Sleep(100 * time.Millisecond)
					stream.Entries = append(stream.Entries, &store.StreamEntry{
						Id:     "1000-0",
						Fields: map[string]string{"key": "value1"},
					})
					stream.LastID = "1000-0"
				}()
			},
			args:    []string{"block", "500", "streams", "mystream", "$"},
			timeout: 2 * time.Second,
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "1000-0") {
					t.Errorf("Should receive the new entry added during blocking, got: %q", result)
				}
			},
		},
		{
			name: "Blocking read timeout - no data",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{},
					LastID:  "0-0",
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args:     []string{"block", "100", "streams", "mystream", "$"},
			expected: "$-1\r\n", // Null response on timeout
			timeout:  1 * time.Second,
		},
		{
			name: "Read with $ special ID",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "1000-0", Fields: map[string]string{"key": "value1"}},
						{Id: "2000-0", Fields: map[string]string{"key": "value2"}},
					},
					LastID: "2000-0",
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args:     []string{"streams", "mystream", "$"},
			expected: "$-1\r\n", // No entries after last ID
		},
		{
			name: "Wrong number of arguments - missing ID",
			setup: func() {
			},
			args:     []string{"streams", "mystream"},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Missing STREAMS keyword",
			setup: func() {
			},
			args:     []string{"mystream", "0-0"},
			expected: "-wrong number of arguments\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeout > 0 {
				// Set a timeout for tests that might block
				done := make(chan bool)
				go func() {
					tt.setup()
					cli := setupTestClient()
					cmd := command.New("xread", tt.args)
					result := cmd.Execute(cli)

					if tt.expected != "" && string(result) != tt.expected {
						t.Errorf("Expected %q, got %q", tt.expected, string(result))
					}

					if tt.validate != nil {
						tt.validate(t, string(result))
					}
					done <- true
				}()

				select {
				case <-done:
					// Test completed
				case <-time.After(tt.timeout):
					t.Fatal("Test timed out")
				}
			} else {
				tt.setup()
				cli := setupTestClient()
				cmd := command.New("xread", tt.args)
				result := cmd.Execute(cli)

				if tt.expected != "" && string(result) != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, string(result))
				}

				if tt.validate != nil {
					tt.validate(t, string(result))
				}
			}
		})
	}
}
