package tests

import (
	"net"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupGeoAddTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestGeoAddCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		expected string
		validate func(t *testing.T)
	}{
		{
			name: "Add single location",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "13.361389", "38.115556", "Palermo"},
			expected: ":1\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("locations")
				if !exists || val.SortedSetData == nil {
					t.Fatal("Expected sorted set to be created")
				}
				if val.SortedSetData.Card() != 1 {
					t.Errorf("Expected 1 member, got %d", val.SortedSetData.Card())
				}
				_, memberExists := val.SortedSetData.GetScore("Palermo")
				if !memberExists {
					t.Error("Expected Palermo to exist in sorted set")
				}
			},
		},
		{
			name: "Add multiple locations",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "13.361389", "38.115556", "Palermo", "15.087269", "37.502669", "Catania"},
			expected: ":2\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("locations")
				if !exists {
					t.Fatal("Expected sorted set to exist")
				}
				if val.SortedSetData.Card() != 2 {
					t.Errorf("Expected 2 members, got %d", val.SortedSetData.Card())
				}
			},
		},
		{
			name: "Update existing location",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoAddTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo"}).Execute(cli)
			},
			args:     []string{"locations", "15.087269", "37.502669", "Palermo"},
			expected: ":0\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("locations")
				if !exists {
					t.Fatal("Expected sorted set to exist")
				}
				if val.SortedSetData.Card() != 1 {
					t.Errorf("Expected 1 member, got %d", val.SortedSetData.Card())
				}
			},
		},
		{
			name: "Add location to existing sorted set",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoAddTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo"}).Execute(cli)
			},
			args:     []string{"locations", "15.087269", "37.502669", "Catania"},
			expected: ":1\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("locations")
				if !exists {
					t.Fatal("Expected sorted set to exist")
				}
				if val.SortedSetData.Card() != 2 {
					t.Errorf("Expected 2 members, got %d", val.SortedSetData.Card())
				}
			},
		},
		{
			name: "Error on invalid longitude - too low",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "-181.0", "38.115556", "Invalid"},
			expected: "-ERR invalid longitude, must be between -180.0000000000 and 180.0000000000\r\n",
		},
		{
			name: "Error on invalid longitude - too high",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "181.0", "38.115556", "Invalid"},
			expected: "-ERR invalid longitude, must be between -180.0000000000 and 180.0000000000\r\n",
		},
		{
			name: "Error on invalid latitude - too low",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "13.361389", "-86.0", "Invalid"},
			expected: "-ERR invalid latitude, must be between -85.0511287800 and 85.0511287800\r\n",
		},
		{
			name: "Error on invalid latitude - too high",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "13.361389", "86.0", "Invalid"},
			expected: "-ERR invalid latitude, must be between -85.0511287800 and 85.0511287800\r\n",
		},
		{
			name: "Error on invalid longitude format",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "invalid", "38.115556", "Palermo"},
			expected: "-ERR value is not a valid float\r\n",
		},
		{
			name: "Error on invalid latitude format",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "13.361389", "invalid", "Palermo"},
			expected: "-ERR value is not a valid float\r\n",
		},
		{
			name:     "Error on wrong number of arguments - too few",
			setup:    func() {},
			args:     []string{"locations", "13.361389", "38.115556"},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name:     "Error on wrong number of arguments - incomplete group",
			setup:    func() {},
			args:     []string{"locations", "13.361389", "38.115556", "Palermo", "15.087269"},
			expected: "-ERR syntax error\r\n",
		},
		{
			name:     "Error on no arguments",
			setup:    func() {},
			args:     []string{},
			expected: "-wrong number of arguments\r\n",
		},
		{
			name: "Error on wrong type - key exists as string",
			setup: func() {
				store.Delete("mystring")
				cli := setupGeoAddTestClient()
				command.New("set", []string{"mystring", "value"}).Execute(cli)
			},
			args:     []string{"mystring", "13.361389", "38.115556", "Palermo"},
			expected: "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
		},
		{
			name: "Add location at longitude boundaries",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "-180.0", "0.0", "West", "180.0", "0.0", "East"},
			expected: ":2\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("locations")
				if !exists {
					t.Fatal("Expected sorted set to exist")
				}
				if val.SortedSetData.Card() != 2 {
					t.Errorf("Expected 2 members, got %d", val.SortedSetData.Card())
				}
			},
		},
		{
			name: "Add location at latitude boundaries",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "0.0", "-85.05112878", "South", "0.0", "85.05112878", "North"},
			expected: ":2\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("locations")
				if !exists {
					t.Fatal("Expected sorted set to exist")
				}
				if val.SortedSetData.Card() != 2 {
					t.Errorf("Expected 2 members, got %d", val.SortedSetData.Card())
				}
			},
		},
		{
			name: "Add location at equator and prime meridian",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "0.0", "0.0", "Origin"},
			expected: ":1\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("locations")
				if !exists {
					t.Fatal("Expected sorted set to exist")
				}
				_, memberExists := val.SortedSetData.GetScore("Origin")
				if !memberExists {
					t.Error("Expected Origin to exist in sorted set")
				}
			},
		},
		{
			name: "Add multiple locations with same member name overwrites",
			setup: func() {
				store.Delete("locations")
			},
			args:     []string{"locations", "13.361389", "38.115556", "City", "15.087269", "37.502669", "City"},
			expected: ":1\r\n",
			validate: func(t *testing.T) {
				val, exists := store.Get("locations")
				if !exists {
					t.Fatal("Expected sorted set to exist")
				}
				if val.SortedSetData.Card() != 1 {
					t.Errorf("Expected 1 member (overwritten), got %d", val.SortedSetData.Card())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupGeoAddTestClient()
			cmd := command.New("geoadd", tt.args)
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
