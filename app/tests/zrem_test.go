package tests

import (
	"net"
	"testing"

	command "github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func setupZRemTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestZRemCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
	}{
		{
			name: "Remove single member",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRemTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three"}).Execute(cli)
			},
			args:     []string{"myzset", "two"},
			expected: ":1\r\n",
		},
		{
			name: "Remove multiple members",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRemTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three", "4", "four"}).Execute(cli)
			},
			args:     []string{"myzset", "one", "three", "four"},
			expected: ":3\r\n",
		},
		{
			name: "Remove non-existent member",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRemTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two"}).Execute(cli)
			},
			args:     []string{"myzset", "nonexistent"},
			expected: ":0\r\n",
		},
		{
			name: "Remove mix of existing and non-existent members",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRemTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three"}).Execute(cli)
			},
			args:     []string{"myzset", "one", "nonexistent", "two"},
			expected: ":2\r\n",
		},
		{
			name: "Remove from non-existent key",
			setup: func() {
				store.Delete("nonexistent")
			},
			args:     []string{"nonexistent", "member"},
			expected: ":0\r\n",
		},
		{
			name: "Remove all members",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRemTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two"}).Execute(cli)
			},
			args:     []string{"myzset", "one", "two"},
			expected: ":2\r\n",
		},
		{
			name: "Error on wrong number of arguments - no args",
			setup: func() {
			},
			args:     []string{},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Error on wrong number of arguments - only key",
			setup: func() {
			},
			args:     []string{"myzset"},
			expected: "-wrong number of arguments\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupZRemTestClient()
			cmd := command.New("zrem", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}
