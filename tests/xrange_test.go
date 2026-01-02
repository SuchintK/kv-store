package tests

import (
	"strings"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/store"
)

func TestXRangeCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
		validate func(*testing.T, string)
	}{
		{
			name: "Range query with explicit sequence numbers",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "1000-0", Fields: map[string]string{"key": "value1"}},
						{Id: "2000-0", Fields: map[string]string{"key": "value2"}},
						{Id: "3000-0", Fields: map[string]string{"key": "value3"}},
					},
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args: []string{"mystream", "1000-0", "2000-0"},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "1000-0") || !strings.Contains(result, "2000-0") {
					t.Errorf("Expected range from 1000-0 to 2000-0, got %q", result)
				}
				if strings.Contains(result, "3000-0") {
					t.Error("Should not include 3000-0")
				}
			},
		},
		{
			name: "Range query with missing sequence numbers",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "1000-0", Fields: map[string]string{"key": "value1"}},
						{Id: "1000-1", Fields: map[string]string{"key": "value2"}},
						{Id: "1000-2", Fields: map[string]string{"key": "value3"}},
						{Id: "2000-0", Fields: map[string]string{"key": "value4"}},
					},
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args: []string{"mystream", "1000", "1000"},
			validate: func(t *testing.T, result string) {
				// Start defaults to 1000-0, end defaults to 1000-max
				if !strings.Contains(result, "1000-0") || !strings.Contains(result, "1000-1") || !strings.Contains(result, "1000-2") {
					t.Errorf("Expected all entries with timestamp 1000, got %q", result)
				}
				if strings.Contains(result, "2000-0") {
					t.Error("Should not include 2000-0")
				}
			},
		},
		{
			name: "Range query with - (minimum)",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "1000-0", Fields: map[string]string{"key": "value1"}},
						{Id: "2000-0", Fields: map[string]string{"key": "value2"}},
					},
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args: []string{"mystream", "-", "1500-0"},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "1000-0") {
					t.Error("Should include 1000-0")
				}
				if strings.Contains(result, "2000-0") {
					t.Error("Should not include 2000-0")
				}
			},
		},
		{
			name: "Range query with + (maximum)",
			setup: func() {
				stream := &store.Stream{
					Entries: []*store.StreamEntry{
						{Id: "1000-0", Fields: map[string]string{"key": "value1"}},
						{Id: "2000-0", Fields: map[string]string{"key": "value2"}},
						{Id: "3000-0", Fields: map[string]string{"key": "value3"}},
					},
				}
				store.Set("mystream", &store.Value{StreamData: stream})
			},
			args: []string{"mystream", "1500-0", "+"},
			validate: func(t *testing.T, result string) {
				if strings.Contains(result, "1000-0") {
					t.Error("Should not include 1000-0")
				}
				if !strings.Contains(result, "2000-0") || !strings.Contains(result, "3000-0") {
					t.Error("Should include all entries after 1500-0")
				}
			},
		},
		{
			name: "Non-existent stream",
			setup: func() {
				// No setup needed
			},
			args:     []string{"nonexistent", "-", "+"},
			expected: "*0\r\n",
		},
		{
			name: "Wrong number of arguments",
			setup: func() {
			},
			args:     []string{"mystream"},
			expected: "-wrong number of arguments\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupTestClient()
			cmd := command.New("xrange", tt.args)
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
