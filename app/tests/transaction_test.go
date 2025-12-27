package tests

import (
	"strings"
	"testing"

	command "github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/resp/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func TestMultiCommand(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*client.Client)
		args        []string
		expected    string
		shouldError bool
	}{
		{
			name: "Start transaction successfully",
			setup: func(cli *client.Client) {
				// Client is not in transaction
			},
			args:        []string{},
			expected:    "+OK\r\n",
			shouldError: false,
		},
		{
			name: "Error on nested MULTI",
			setup: func(cli *client.Client) {
				cli.StartTransaction()
			},
			args:        []string{},
			expected:    "-ERR MULTI calls can not be nested\r\n",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := setupTestClient()
			tt.setup(cli)
			cmd := command.New("multi", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}

			if !tt.shouldError && !cli.IsInTransaction() {
				t.Error("Client should be in transaction after MULTI")
			}
		})
	}
}

func TestExecCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*client.Client)
		args     []string
		expected string
		validate func(*testing.T, *client.Client, string)
	}{
		{
			name: "Execute empty transaction",
			setup: func(cli *client.Client) {
				cli.StartTransaction()
			},
			args:     []string{},
			expected: "*0\r\n",
			validate: func(t *testing.T, cli *client.Client, result string) {
				if cli.IsInTransaction() {
					t.Error("Client should not be in transaction after EXEC")
				}
			},
		},
		{
			name: "Execute transaction with SET commands",
			setup: func(cli *client.Client) {
				cli.StartTransaction()
				cli.QueueCommand("set", []string{"key1", "value1"})
				cli.QueueCommand("set", []string{"key2", "value2"})
			},
			args: []string{},
			validate: func(t *testing.T, cli *client.Client, result string) {
				if !strings.Contains(result, "+OK") {
					t.Error("SET commands should return OK")
				}
				val1, exists1 := store.Get("key1")
				val2, exists2 := store.Get("key2")
				if !exists1 || !exists2 || val1.Data != "value1" || val2.Data != "value2" {
					t.Error("SET commands in transaction were not executed")
				}
			},
		},
		{
			name: "Execute transaction with INCR commands",
			setup: func(cli *client.Client) {
				store.Set("counter", &store.Value{Data: "10"})
				cli.StartTransaction()
				cli.QueueCommand("incr", []string{"counter"})
				cli.QueueCommand("incr", []string{"counter"})
			},
			args: []string{},
			validate: func(t *testing.T, cli *client.Client, result string) {
				if !strings.Contains(result, ":11\r\n") || !strings.Contains(result, ":12\r\n") {
					t.Errorf("Expected :11 and :12 in result, got %q", result)
				}
				val, _ := store.Get("counter")
				if val.Data != "12" {
					t.Errorf("Expected counter to be 12, got %s", val.Data)
				}
			},
		},
		{
			name: "Handle errors within transaction",
			setup: func(cli *client.Client) {
				store.Set("notanumber", &store.Value{Data: "abc"})
				cli.StartTransaction()
				cli.QueueCommand("incr", []string{"validkey"})
				cli.QueueCommand("incr", []string{"notanumber"}) // This will error
				cli.QueueCommand("incr", []string{"validkey"})   // Should still execute
			},
			args: []string{},
			validate: func(t *testing.T, cli *client.Client, result string) {
				if !strings.Contains(result, ":1\r\n") {
					t.Error("First INCR should succeed")
				}
				if !strings.Contains(result, "-ERR") {
					t.Error("Second INCR should fail with error")
				}
				if !strings.Contains(result, ":2\r\n") {
					t.Error("Third INCR should still execute")
				}
			},
		},
		{
			name: "Error when not in transaction",
			setup: func(cli *client.Client) {
				// Client is not in transaction
			},
			args:     []string{},
			expected: "-ERR EXEC without MULTI\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := setupTestClient()
			tt.setup(cli)

			cmd := command.New("exec", tt.args)
			result := cmd.Execute(cli)

			if tt.expected != "" && string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}

			if tt.validate != nil {
				tt.validate(t, cli, string(result))
			}
		})
	}
}

func TestDiscardCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*client.Client)
		args     []string
		expected string
		validate func(*testing.T, *client.Client)
	}{
		{
			name: "Discard transaction successfully",
			setup: func(cli *client.Client) {
				cli.StartTransaction()
				cli.QueueCommand("set", []string{"key1", "value1"})
			},
			args:     []string{},
			expected: "+OK\r\n",
			validate: func(t *testing.T, cli *client.Client) {
				if cli.IsInTransaction() {
					t.Error("Client should not be in transaction after DISCARD")
				}
				if len(cli.GetQueuedCommands()) != 0 {
					t.Error("Queued commands should be cleared")
				}
			},
		},
		{
			name: "Error when not in transaction",
			setup: func(cli *client.Client) {
				// Client is not in transaction
			},
			args:     []string{},
			expected: "-ERR DISCARD without MULTI\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := setupTestClient()
			tt.setup(cli)

			cmd := command.New("discard", tt.args)
			result := cmd.Execute(cli)

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}

			if tt.validate != nil {
				tt.validate(t, cli)
			}
		})
	}
}
