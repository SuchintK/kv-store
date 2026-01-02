package tests

import (
	"net"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupZScoreTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestZScoreCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
	}{
		{
			name: "Get score of existing member",
			setup: func() {
				store.Delete("myzset")
				cli := setupZScoreTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three"}).Execute(cli)
			},
			args:     []string{"myzset", "two"},
			expected: "$1\r\n2\r\n",
		},
		{
			name: "Get score of member with decimal score",
			setup: func() {
				store.Delete("myzset")
				cli := setupZScoreTestClient()
				command.New("zadd", []string{"myzset", "1.5", "one", "2.75", "two"}).Execute(cli)
			},
			args:     []string{"myzset", "two"},
			expected: "$4\r\n2.75\r\n",
		},
		{
			name: "Get score of member with negative score",
			setup: func() {
				store.Delete("myzset")
				cli := setupZScoreTestClient()
				command.New("zadd", []string{"myzset", "-5", "negative", "10", "positive"}).Execute(cli)
			},
			args:     []string{"myzset", "negative"},
			expected: "$2\r\n-5\r\n",
		},
		{
			name: "Non-existent member returns null",
			setup: func() {
				store.Delete("myzset")
				cli := setupZScoreTestClient()
				command.New("zadd", []string{"myzset", "1", "one"}).Execute(cli)
			},
			args:     []string{"myzset", "nonexistent"},
			expected: "$-1\r\n",
		},
		{
			name: "Non-existent key returns null",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupZScoreTestClient()
			cmd := command.New("zscore", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}
