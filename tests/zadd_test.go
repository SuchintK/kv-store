package tests

import (
	"net"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupZAddTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestZAddCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
		validate func(t *testing.T)
	}{
		{
			name: "Add single member to new sorted set",
			setup: func() {
				store.Delete("myzset")
			},
			args:     []string{"myzset", "1.5", "member1"},
			expected: ":1\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("myzset")
				if !exists || val.SortedSetData == nil {
					t.Fatal("Expected sorted set to be created")
				}
				if val.SortedSetData.Card() != 1 {
					t.Errorf("Expected 1 member, got %d", val.SortedSetData.Card())
				}
				score, exists := val.SortedSetData.GetScore("member1")
				if !exists {
					t.Fatal("Expected member1 to exist")
				}
				if score != 1.5 {
					t.Errorf("Expected score 1.5, got %f", score)
				}
			},
		},
		{
			name: "Add multiple members",
			setup: func() {
				store.Delete("myzset")
			},
			args:     []string{"myzset", "1", "one", "2", "two", "3", "three"},
			expected: ":3\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("myzset")
				if !exists {
					t.Fatal("Expected key to exist")
				}
				if val.SortedSetData.Card() != 3 {
					t.Errorf("Expected 3 members, got %d", val.SortedSetData.Card())
				}
				// Verify each member and score
				score1, exists := val.SortedSetData.GetScore("one")
				if !exists || score1 != 1 {
					t.Errorf("Expected one with score 1, got %f", score1)
				}
				score2, exists := val.SortedSetData.GetScore("two")
				if !exists || score2 != 2 {
					t.Errorf("Expected two with score 2, got %f", score2)
				}
				score3, exists := val.SortedSetData.GetScore("three")
				if !exists || score3 != 3 {
					t.Errorf("Expected three with score 3, got %f", score3)
				}
			},
		},
		{
			name: "Update existing member score",
			setup: func() {
				store.Delete("myzset")
				cli := setupZAddTestClient()
				command.New("zadd", []string{"myzset", "1", "member"}).Execute(cli)
			},
			args:     []string{"myzset", "2", "member"},
			expected: ":0\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("myzset")
				if !exists {
					t.Fatal("Expected key to exist")
				}
				score, exists := val.SortedSetData.GetScore("member")
				if !exists {
					t.Fatal("Expected member to exist")
				}
				if score != 2 {
					t.Errorf("Expected score 2, got %f", score)
				}
			},
		},
		{
			name: "Add member with same score and member",
			setup: func() {
				store.Delete("myzset")
				cli := setupZAddTestClient()
				command.New("zadd", []string{"myzset", "1", "member1"}).Execute(cli)
			},
			args:     []string{"myzset", "1", "member1"},
			expected: ":0\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("myzset")
				if !exists {
					t.Fatal("Expected key to exist")
				}
				score, exists := val.SortedSetData.GetScore("member1")
				if !exists {
					t.Fatal("Expected member to exist")
				}
				if score != 1 {
					t.Errorf("Expected score 1, got %f", score)
				}
				if val.SortedSetData.Card() != 1 {
					t.Errorf("Expected 1 member, got %d", val.SortedSetData.Card())
				}
			},
		},
		{
			name: "Add members with negative and decimal scores",
			setup: func() {
				store.Delete("myzset")
			},
			args:     []string{"myzset", "-10.5", "negative", "0", "zero", "99.99", "decimal"},
			expected: ":3\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("myzset")
				if !exists {
					t.Fatal("Expected key to exist")
				}
				if val.SortedSetData.Card() != 3 {
					t.Errorf("Expected 3 members, got %d", val.SortedSetData.Card())
				}
				// Verify negative score
				score1, exists := val.SortedSetData.GetScore("negative")
				if !exists || score1 != -10.5 {
					t.Errorf("Expected negative with score -10.5, got %f", score1)
				}
				// Verify zero score
				score2, exists := val.SortedSetData.GetScore("zero")
				if !exists || score2 != 0 {
					t.Errorf("Expected zero with score 0, got %f", score2)
				}
				// Verify decimal score
				score3, exists := val.SortedSetData.GetScore("decimal")
				if !exists || score3 != 99.99 {
					t.Errorf("Expected decimal with score 99.99, got %f", score3)
				}
			},
		},
		{
			name: "Error on invalid score",
			setup: func() {
				store.Delete("myzset")
			},
			args:     []string{"myzset", "invalid", "member"},
			expected: "-ERR value is not a valid float\r\n",
		},
		{
			name: "Error on too few arguments",
			setup: func() {
			},
			args:     []string{"myzset"},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Error on odd number of score-member pairs",
			setup: func() {
			},
			args:     []string{"myzset", "1", "member1", "2"},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Error on wrong type",
			setup: func() {
				store.Delete("mystring")
				cli := setupZAddTestClient()
				command.New("set", []string{"mystring", "value"}).Execute(cli)
			},
			args:     []string{"mystring", "1", "member"},
			expected: "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupZAddTestClient()
			cmd := command.New("zadd", tt.args)
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
