package tests

import (
	"net"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupZRangeTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestZRangeCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
	}{
		{
			name: "Get range without scores",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRangeTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three", "4", "four", "5", "five"}).Execute(cli)
			},
			args:     []string{"myzset", "0", "2"},
			expected: "*3\r\n$3\r\none\r\n$3\r\ntwo\r\n$5\r\nthree\r\n",
		},
		{
			name: "Get range with WITHSCORES",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRangeTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two"}).Execute(cli)
			},
			args:     []string{"myzset", "0", "1", "withscores"},
			expected: "*4\r\n$3\r\none\r\n$1\r\n1\r\n$3\r\ntwo\r\n$1\r\n2\r\n",
		},
		{
			name: "Get all elements with negative stop",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRangeTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three"}).Execute(cli)
			},
			args:     []string{"myzset", "0", "-1"},
			expected: "*3\r\n$3\r\none\r\n$3\r\ntwo\r\n$5\r\nthree\r\n",
		},
		{
			name: "Get range with negative start",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRangeTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three", "4", "four"}).Execute(cli)
			},
			args:     []string{"myzset", "-3", "-1"},
			expected: "*3\r\n$3\r\ntwo\r\n$5\r\nthree\r\n$4\r\nfour\r\n",
		},
		{
			name: "Get single element",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRangeTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two", "3", "three"}).Execute(cli)
			},
			args:     []string{"myzset", "1", "1"},
			expected: "*1\r\n$3\r\ntwo\r\n",
		},
		{
			name: "Out of range returns empty array",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRangeTestClient()
				command.New("zadd", []string{"myzset", "1", "one"}).Execute(cli)
			},
			args:     []string{"myzset", "10", "20"},
			expected: "*0\r\n",
		},
		{
			name: "Start greater than stop returns empty array",
			setup: func() {
				store.Delete("myzset")
				cli := setupZRangeTestClient()
				command.New("zadd", []string{"myzset", "1", "one", "2", "two"}).Execute(cli)
			},
			args:     []string{"myzset", "2", "1"},
			expected: "*0\r\n",
		},
		{
			name: "Non-existent key returns empty array",
			setup: func() {
				store.Delete("nonexistent")
			},
			args:     []string{"nonexistent", "0", "1"},
			expected: "*0\r\n",
		},
		{
			name: "Error on wrong number of arguments - too few",
			setup: func() {
			},
			args:     []string{"myzset", "0"},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Error on wrong number of arguments - no args",
			setup: func() {
			},
			args:     []string{},
			expected: "-wrong number of arguments\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupZRangeTestClient()
			cmd := command.New("zrange", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}
