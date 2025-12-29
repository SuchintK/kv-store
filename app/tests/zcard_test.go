package tests

import (
	"net"
	"testing"

	command "github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func setupZCardTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestZCardCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
	}{
		{
			name: "Get cardinality of sorted set",
			setup: func() {
				store.Delete("myzset")
				cli := setupZCardTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three", "4", "four", "5", "five"}).Execute(cli)
			},
			args:     []string{"myzset"},
			expected: ":5\r\n",
		},
		{
			name: "Get cardinality of single element set",
			setup: func() {
				store.Delete("myzset")
				cli := setupZCardTestClient()
				command.New("zadd", []string{"myzset", "1", "one"}).Execute(cli)
			},
			args:     []string{"myzset"},
			expected: ":1\r\n",
		},
		{
			name: "Non-existent key returns zero",
			setup: func() {
				store.Delete("nonexistent")
			},
			args:     []string{"nonexistent"},
			expected: ":0\r\n",
		},
		{
			name: "Empty sorted set returns zero",
			setup: func() {
				store.Delete("myzset")
				cli := setupZCardTestClient()
				command.New("zadd", []string{"myzset", "1", "one"}).Execute(cli)
				command.New("zrem", []string{"myzset", "one"}).Execute(cli)
			},
			args:     []string{"myzset"},
			expected: ":0\r\n",
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
			args:     []string{"myzset", "extra"},
			expected: "-wrong number of arguments\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupZCardTestClient()
			cmd := command.New("zcard", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}
