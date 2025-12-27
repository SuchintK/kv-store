package tests

import (
	"net"
	"testing"

	command "github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func setupTestClient() *client.Client {
	// Create a mock connection for testing
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestIncrCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
	}{
		{
			name: "Increment non-existent key",
			setup: func() {
				// Key doesn't exist, should initialize to 0 and return 1
			},
			args:     []string{"counter"},
			expected: ":1\r\n",
		},
		{
			name: "Increment existing integer",
			setup: func() {
				store.Set("counter", &store.Value{Data: "5"})
			},
			args:     []string{"counter"},
			expected: ":6\r\n",
		},
		{
			name: "Increment negative integer",
			setup: func() {
				store.Set("counter", &store.Value{Data: "-10"})
			},
			args:     []string{"counter"},
			expected: ":-9\r\n",
		},
		{
			name: "Error on non-integer value",
			setup: func() {
				store.Set("mykey", &store.Value{Data: "notanumber"})
			},
			args:     []string{"mykey"},
			expected: "-ERR value is not an integer or out of range\r\n",
		},
		{
			name: "Error on wrong number of arguments - no args",
			setup: func() {
			},
			args:     []string{},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Error on wrong number of arguments - too many",
			setup: func() {
			},
			args:     []string{"key1", "key2"},
			expected: "-wrong number of arguments\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupTestClient()
			cmd := command.New("incr", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}
