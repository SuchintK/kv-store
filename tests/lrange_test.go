package tests

import (
	"net"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupLRangeTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestLRangeCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
	}{
		{
			name: "Get range from start to middle",
			setup: func() {
				store.Delete("mylist")
				cli := setupLRangeTestClient()
				command.New("rpush", []string{"mylist", "one", "two", "three", "four", "five"}).Execute(cli)
			},
			args:     []string{"mylist", "0", "2"},
			expected: "*3\r\n$3\r\none\r\n$3\r\ntwo\r\n$5\r\nthree\r\n",
		},
		{
			name: "Get all elements with negative stop",
			setup: func() {
				store.Delete("mylist")
				cli := setupLRangeTestClient()
				command.New("rpush", []string{"mylist", "one", "two", "three"}).Execute(cli)
			},
			args:     []string{"mylist", "0", "-1"},
			expected: "*3\r\n$3\r\none\r\n$3\r\ntwo\r\n$5\r\nthree\r\n",
		},
		{
			name: "Get range with negative start",
			setup: func() {
				store.Delete("mylist")
				cli := setupLRangeTestClient()
				command.New("rpush", []string{"mylist", "one", "two", "three", "four"}).Execute(cli)
			},
			args:     []string{"mylist", "-3", "-1"},
			expected: "*3\r\n$3\r\ntwo\r\n$5\r\nthree\r\n$4\r\nfour\r\n",
		},
		{
			name: "Get single element",
			setup: func() {
				store.Delete("mylist")
				cli := setupLRangeTestClient()
				command.New("rpush", []string{"mylist", "one", "two", "three"}).Execute(cli)
			},
			args:     []string{"mylist", "1", "1"},
			expected: "*1\r\n$3\r\ntwo\r\n",
		},
		{
			name: "Out of range returns empty array",
			setup: func() {
				store.Delete("mylist")
				cli := setupLRangeTestClient()
				command.New("rpush", []string{"mylist", "one"}).Execute(cli)
			},
			args:     []string{"mylist", "10", "20"},
			expected: "*0\r\n",
		},
		{
			name: "Start greater than stop returns empty array",
			setup: func() {
				store.Delete("mylist")
				cli := setupLRangeTestClient()
				command.New("rpush", []string{"mylist", "one", "two"}).Execute(cli)
			},
			args:     []string{"mylist", "2", "1"},
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
			name: "Get last element with negative indices",
			setup: func() {
				store.Delete("mylist")
				cli := setupLRangeTestClient()
				command.New("rpush", []string{"mylist", "one", "two", "three"}).Execute(cli)
			},
			args:     []string{"mylist", "-1", "-1"},
			expected: "*1\r\n$5\r\nthree\r\n",
		},
		{
			name: "Error on wrong type",
			setup: func() {
				store.Delete("mystring")
				cli := setupLRangeTestClient()
				command.New("set", []string{"mystring", "value"}).Execute(cli)
			},
			args:     []string{"mystring", "0", "1"},
			expected: "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
		},
		{
			name: "Error on wrong number of arguments - too few",
			setup: func() {
			},
			args:     []string{"mylist", "0"},
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
			cli := setupLRangeTestClient()
			cmd := command.New("lrange", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}
