package tests

import (
	"net"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupLPushTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestLPushCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
		validate func(t *testing.T)
	}{
		{
			name: "Push single element to new list",
			setup: func() {
				store.Delete("mylist")
			},
			args:     []string{"mylist", "world"},
			expected: ":1\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("mylist")
				if !exists || val.ListData == nil {
					t.Fatal("Expected list to be created")
				}
				if len(val.ListData) != 1 {
					t.Errorf("Expected 1 element, got %d", len(val.ListData))
				}
				if val.ListData[0] != "world" {
					t.Errorf("Expected 'world', got %s", val.ListData[0])
				}
			},
		},
		{
			name: "Push multiple elements",
			setup: func() {
				store.Delete("mylist")
			},
			args:     []string{"mylist", "one", "two", "three"},
			expected: ":3\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("mylist")
				if !exists {
					t.Fatal("Expected list to exist")
				}
				if len(val.ListData) != 3 {
					t.Errorf("Expected 3 elements, got %d", len(val.ListData))
				}
				if val.ListData[0] != "three" || val.ListData[1] != "two" || val.ListData[2] != "one" {
					t.Errorf("Expected [three, two, one], got %v", val.ListData)
				}
			},
		},
		{
			name: "Push to existing list",
			setup: func() {
				store.Delete("mylist")
				cli := setupLPushTestClient()
				command.New("lpush", []string{"mylist", "hello"}).Execute(cli)
			},
			args:     []string{"mylist", "world"},
			expected: ":2\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("mylist")
				if !exists {
					t.Fatal("Expected list to exist")
				}
				if len(val.ListData) != 2 {
					t.Errorf("Expected 2 elements, got %d", len(val.ListData))
				}
				if val.ListData[0] != "world" || val.ListData[1] != "hello" {
					t.Errorf("Expected [world, hello], got %v", val.ListData)
				}
			},
		},
		{
			name: "Error on wrong type",
			setup: func() {
				store.Delete("mystring")
				cli := setupLPushTestClient()
				command.New("set", []string{"mystring", "value"}).Execute(cli)
			},
			args:     []string{"mystring", "element"},
			expected: "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
		},
		{
			name: "Error on wrong number of arguments - no elements",
			setup: func() {
			},
			args:     []string{"mylist"},
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
			cli := setupLPushTestClient()
			cmd := command.New("lpush", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}

			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}
