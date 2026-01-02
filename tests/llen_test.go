package tests

import (
	"net"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupLLenTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestLLenCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
	}{
		{
			name: "Get length of list with elements",
			setup: func() {
				store.Delete("mylist")
				cli := setupLLenTestClient()
				command.New("rpush", []string{"mylist", "one", "two", "three"}).Execute(cli)
			},
			args:     []string{"mylist"},
			expected: ":3\r\n",
		},
		{
			name: "Get length of non-existent key returns zero",
			setup: func() {
				store.Delete("nonexistent")
			},
			args:     []string{"nonexistent"},
			expected: ":0\r\n",
		},
		{
			name: "Get length of empty list returns zero",
			setup: func() {
				store.Set("emptylist", &store.Value{ListData: []string{}})
			},
			args:     []string{"emptylist"},
			expected: ":0\r\n",
		},
		{
			name: "Get length of single element list",
			setup: func() {
				store.Delete("mylist")
				cli := setupLLenTestClient()
				command.New("rpush", []string{"mylist", "only"}).Execute(cli)
			},
			args:     []string{"mylist"},
			expected: ":1\r\n",
		},
		{
			name: "Error on wrong type",
			setup: func() {
				store.Delete("mystring")
				cli := setupLLenTestClient()
				command.New("set", []string{"mystring", "value"}).Execute(cli)
			},
			args:     []string{"mystring"},
			expected: "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
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
			args:     []string{"mylist", "extra"},
			expected: "-wrong number of arguments\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupLLenTestClient()
			cmd := command.New("llen", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}
