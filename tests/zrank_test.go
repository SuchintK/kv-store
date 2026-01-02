package tests

import (
	"net"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupZRankTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestZRankCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
	}{
		{
			name: "Get rank of first member",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRankTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three"}).Execute(cli)
			},
			args:     []string{"myzset", "one"},
			expected: ":0\r\n",
		},
		{
			name: "Get rank of middle member",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRankTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three"}).Execute(cli)
			},
			args:     []string{"myzset", "two"},
			expected: ":1\r\n",
		},
		{
			name: "Get rank of last member",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRankTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three"}).Execute(cli)
			},
			args:     []string{"myzset", "three"},
			expected: ":2\r\n",
		},
		{
			name: "Get rank of non-existent member",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRankTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two"}).Execute(cli)
			},
			args:     []string{"myzset", "nonexistent"},
			expected: "$-1\r\n",
		},
		{
			name: "Get rank from non-existent key",
			setup: func() {
				store.Delete("nonexistent")
			},
			args:     []string{"nonexistent", "member"},
			expected: "$-1\r\n",
		},
		{
			name: "Error on wrong number of arguments - too few",
			setup: func() {
			},
			args:     []string{"myzset"},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Error on wrong number of arguments - too many",
			setup: func() {
			},
			args:     []string{"myzset", "member", "extra"},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Get rank with same score, lexicographically different members",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRankTestClient()
				// All have score 5, but different names
				command.New("zadd", []string{"myzset", "5", "apple", "5", "banana", "5", "cherry"}).Execute(cli)
			},
			args:     []string{"myzset", "banana"},
			expected: ":1\r\n",
		},
		{
			name: "Get rank with same score, lexicographically first member",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRankTestClient()
				command.New("zadd", []string{"myzset", "5", "apple", "5", "banana", "5", "cherry"}).Execute(cli)
			},
			args:     []string{"myzset", "apple"},
			expected: ":0\r\n",
		},
		{
			name: "Get rank with same score, lexicographically last member",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRankTestClient()
				command.New("zadd", []string{"myzset", "5", "apple", "5", "banana", "5", "cherry"}).Execute(cli)
			},
			args:     []string{"myzset", "cherry"},
			expected: ":2\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupZRankTestClient()
			cmd := command.New("zrank", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}
