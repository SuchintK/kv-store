package tests

import (
	"net"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupRPopTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestRPopCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
		validate func(t *testing.T)
	}{
		{
			name: "Pop from list with elements",
			setup: func() {
				store.Delete("mylist")
				cli := setupRPopTestClient()
				command.New("rpush", []string{"mylist", "one", "two", "three"}).Execute(cli)
			},
			args:     []string{"mylist"},
			expected: "$5\r\nthree\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("mylist")
				if !exists {
					t.Fatal("Expected list to exist")
				}
				if len(val.ListData) != 2 {
					t.Errorf("Expected 2 elements remaining, got %d", len(val.ListData))
				}
				if val.ListData[1] != "two" {
					t.Errorf("Expected last element to be 'two', got %s", val.ListData[1])
				}
			},
		},
		{
			name: "Pop last element deletes key",
			setup: func() {
				store.Delete("mylist")
				cli := setupRPopTestClient()
				command.New("rpush", []string{"mylist", "only"}).Execute(cli)
			},
			args:     []string{"mylist"},
			expected: "$4\r\nonly\r\n",
			validate: func(t *testing.T) {
				_, exists := store.Get("mylist")
				if exists {
					t.Error("Expected key to be deleted after popping last element")
				}
			},
		},
		{
			name: "Pop from non-existent key returns null",
			setup: func() {
				store.Delete("nonexistent")
			},
			args:     []string{"nonexistent"},
			expected: "$-1\r\n",
		},
		{
			name: "Pop from empty list returns null",
			setup: func() {
				store.Set("emptylist", &store.Value{ListData: []string{}})
			},
			args:     []string{"emptylist"},
			expected: "$-1\r\n",
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
			cli := setupRPopTestClient()
			cmd := command.New("rpop", tt.args)
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
