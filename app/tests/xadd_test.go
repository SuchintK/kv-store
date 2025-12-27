package tests

import (
	"strings"
	"testing"

	command "github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func TestXAddCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
		validate func(*testing.T, string)
	}{
		{
			name: "Add entry with auto-generated ID (*)",
			setup: func() {
				// Clean slate
			},
			args: []string{"mystream", "*", "field1", "value1", "field2", "value2"},
			validate: func(t *testing.T, result string) {
				// Result should be a bulk string with the generated ID
				if !strings.HasPrefix(result, "$") {
					t.Errorf("Expected bulk string response, got %q", result)
				}
				// Verify stream was created
				val, exists := store.Get("mystream")
				if !exists || val.StreamData == nil {
					t.Error("Stream should be created")
				}
				if len(val.StreamData.Entries) != 1 {
					t.Errorf("Expected 1 entry, got %d", len(val.StreamData.Entries))
				}
			},
		},
		{
			name: "Add entry with explicit ID",
			setup: func() {
				stream := &store.Stream{
					Entries:       []*store.StreamEntry{},
					LastTimestamp: 0,
					LastSequence:  0,
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args:     []string{"mystream", "1000-0", "name", "Alice", "age", "30"},
			expected: "$6\r\n1000-0\r\n",
			validate: func(t *testing.T, result string) {
				val, _ := store.Get("mystream")
				if len(val.StreamData.Entries) != 1 {
					t.Error("Should have 1 entry")
				}
				entry := val.StreamData.Entries[0]
				if entry.Id != "1000-0" {
					t.Errorf("Expected ID 1000-0, got %s", entry.Id)
				}
				if entry.Fields["name"] != "Alice" || entry.Fields["age"] != "30" {
					t.Error("Fields not stored correctly")
				}
			},
		},
		{
			name: "Add multiple entries with increasing IDs",
			setup: func() {
				stream := &store.Stream{
					Entries:       []*store.StreamEntry{},
					LastTimestamp: 0,
					LastSequence:  0,
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args: []string{"mystream", "1000-0", "msg", "first"},
			validate: func(t *testing.T, result string) {
				// Add second entry
				cli := setupTestClient()
				cmd2 := command.New("xadd", []string{"mystream", "2000-0", "msg", "second"})
				cmd2.Execute(cli)

				val, _ := store.Get("mystream")
				if len(val.StreamData.Entries) != 2 {
					t.Errorf("Expected 2 entries, got %d", len(val.StreamData.Entries))
				}
			},
		},
		{
			name: "Error on duplicate or lower ID",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "2000-0", Fields: map[string]string{"key": "value"}},
					},
					LastTimestamp: 2000,
					LastSequence:  0,
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args:     []string{"mystream", "1000-0", "field", "value"},
			expected: "-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n",
		},
		{
			name: "Error on equal ID",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "1000-0", Fields: map[string]string{"key": "value"}},
					},
					LastTimestamp: 1000,
					LastSequence:  0,
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args:     []string{"mystream", "1000-0", "field", "value"},
			expected: "-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n",
		},
		{
			name: "Error on same timestamp but lower sequence",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "1000-5", Fields: map[string]string{"key": "value"}},
					},
					LastTimestamp: 1000,
					LastSequence:  5,
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args:     []string{"mystream", "1000-3", "field", "value"},
			expected: "-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n",
		},
		{
			name: "Valid entry with same timestamp but higher sequence",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "1000-0", Fields: map[string]string{"key": "value1"}},
					},
					LastTimestamp: 1000,
					LastSequence:  0,
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args:     []string{"mystream", "1000-1", "field", "value2"},
			expected: "$6\r\n1000-1\r\n",
		},
		{
			name: "Error on invalid ID format - missing dash",
			setup: func() {
				stream := &store.Stream{
					Entries:       []*store.StreamEntry{},
					LastTimestamp: 0,
					LastSequence:  0,
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args:     []string{"mystream", "1000", "field", "value"},
			expected: "-ERR Invalid stream ID specified as stream command argument\r\n",
		},
		{
			name: "Error on invalid ID format - non-numeric timestamp",
			setup: func() {
				stream := &store.Stream{
					Entries:       []*store.StreamEntry{},
					LastTimestamp: 0,
					LastSequence:  0,
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args:     []string{"mystream", "abc-0", "field", "value"},
			expected: "-ERR Invalid stream ID specified as stream command argument\r\n",
		},
		{
			name: "Error on invalid ID format - non-numeric sequence",
			setup: func() {
				stream := &store.Stream{
					Entries:       []*store.StreamEntry{},
					LastTimestamp: 0,
					LastSequence:  0,
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args:     []string{"mystream", "1000-xyz", "field", "value"},
			expected: "-ERR Invalid stream ID specified as stream command argument\r\n",
		},
		{
			name: "Add entry to non-existent stream",
			setup: func() {
				// No setup - stream doesn't exist
			},
			args:     []string{"newstream", "1000-0", "field", "value"},
			expected: "$6\r\n1000-0\r\n",
			validate: func(t *testing.T, result string) {
				val, exists := store.Get("newstream")
				if !exists {
					t.Error("Stream should be created")
				}
				if len(val.StreamData.Entries) != 1 {
					t.Error("Should have 1 entry")
				}
			},
		},
		{
			name: "Error on wrong number of arguments - missing field-value pair",
			setup: func() {
			},
			args:     []string{"mystream", "1000-0", "field"},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Error on wrong number of arguments - too few args",
			setup: func() {
			},
			args:     []string{"mystream", "1000-0"},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Multiple field-value pairs",
			setup: func() {
				stream := &store.Stream{
					Entries:       []*store.StreamEntry{},
					LastTimestamp: 0,
					LastSequence:  0,
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args: []string{"mystream", "1000-0", "name", "Bob", "age", "25", "city", "NYC"},
			validate: func(t *testing.T, result string) {
				val, _ := store.Get("mystream")
				entry := val.StreamData.Entries[0]
				if len(entry.Fields) != 3 {
					t.Errorf("Expected 3 fields, got %d", len(entry.Fields))
				}
				if entry.Fields["name"] != "Bob" || entry.Fields["age"] != "25" || entry.Fields["city"] != "NYC" {
					t.Error("Fields not stored correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupTestClient()
			cmd := command.New("xadd", tt.args)
			result := cmd.Execute(cli)

			if tt.expected != "" && string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}

			if tt.validate != nil {
				tt.validate(t, string(result))
			}
		})
	}
}
