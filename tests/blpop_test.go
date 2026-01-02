package tests

import (
	"net"
	"testing"
	"time"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupBLPopTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestBLPopCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
		validate func(t *testing.T)
	}{
		{
			name: "Pop from existing list immediately",
			setup: func() {
				store.Delete("mylist")
				cli := setupBLPopTestClient()
				command.New("rpush", []string{"mylist", "one", "two", "three"}).Execute(cli)
			},
			args:     []string{"mylist", "1"},
			expected: "*2\r\n$6\r\nmylist\r\n$3\r\none\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("mylist")
				if !exists {
					t.Fatal("Expected list to exist")
				}
				if len(val.ListData) != 2 {
					t.Errorf("Expected 2 elements remaining, got %d", len(val.ListData))
				}
			},
		},
		{
			name: "Pop from first non-empty list among multiple keys",
			setup: func() {
				store.Delete("list1")
				store.Delete("list2")
				cli := setupBLPopTestClient()
				command.New("rpush", []string{"list2", "value"}).Execute(cli)
			},
			args:     []string{"list1", "list2", "1"},
			expected: "*2\r\n$5\r\nlist2\r\n$5\r\nvalue\r\n",
		},
		{
			name: "Timeout on empty list",
			setup: func() {
				store.Delete("emptylist")
			},
			args:     []string{"emptylist", "0.1"},
			expected: "$-1\r\n",
		},
		{
			name: "Pop last element deletes key",
			setup: func() {
				store.Delete("mylist")
				cli := setupBLPopTestClient()
				command.New("rpush", []string{"mylist", "only"}).Execute(cli)
			},
			args:     []string{"mylist", "1"},
			expected: "*2\r\n$6\r\nmylist\r\n$4\r\nonly\r\n",
			validate: func(t *testing.T) {
				_, exists := store.Get("mylist")
				if exists {
					t.Error("Expected key to be deleted after popping last element")
				}
			},
		},
		{
			name: "Error on wrong number of arguments - no timeout",
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
			cli := setupBLPopTestClient()
			cmd := command.New("blpop", tt.args)

			start := time.Now()
			result := cmd.Execute(cli)
			elapsed := time.Since(start)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}

			if tt.name == "Timeout on empty list" {
				if elapsed < 100*time.Millisecond {
					t.Errorf("Expected to wait at least 100ms, but took %v", elapsed)
				}
			}

			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}
