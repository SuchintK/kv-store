package tests

import (
	"net"
	"strings"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupGeoPosTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestGeoPosCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		contains []string
	}{
		{
			name: "Get position of single member",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoPosTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo"}).Execute(cli)
			},
			args:     []string{"locations", "Palermo"},
			contains: []string{"13.361", "38.115"}, // Check prefix due to encoding precision
		},
		{
			name: "Get position of multiple members",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoPosTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo", "15.087269", "37.502669", "Catania"}).Execute(cli)
			},
			args:     []string{"locations", "Palermo", "Catania"},
			contains: []string{"13.361", "38.115", "15.087", "37.502"},
		},
		{
			name: "Get position of non-existent member",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoPosTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo"}).Execute(cli)
			},
			args:     []string{"locations", "NonExistent"},
			contains: []string{"$-1"},
		},
		{
			name: "Get position mix of existing and non-existent members",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoPosTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo"}).Execute(cli)
			},
			args:     []string{"locations", "Palermo", "NonExistent", "Palermo"},
			contains: []string{"13.361", "$-1"}, // Check prefix due to encoding precision
		},
		{
			name: "Get position from non-existent key",
			setup: func() {
				store.Delete("nonexistent")
			},
			args:     []string{"nonexistent", "member1", "member2"},
			contains: []string{"$-1", "$-1"},
		},
		{
			name:     "Error on wrong number of arguments - no members",
			setup:    func() {},
			args:     []string{"locations"},
			contains: []string{"-wrong number of arguments"},
		},
		{
			name:     "Error on no arguments",
			setup:    func() {},
			args:     []string{},
			contains: []string{"-wrong number of arguments"},
		},
		{
			name: "Error on wrong type - key exists as string",
			setup: func() {
				store.Delete("mystring")
				cli := setupGeoPosTestClient()
				command.New("set", []string{"mystring", "value"}).Execute(cli)
			},
			args:     []string{"mystring", "member"},
			contains: []string{"-WRONGTYPE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupGeoPosTestClient()
			cmd := command.New("geopos", tt.args)
			result := cmd.Execute(cli)
			resultStr := string(result)

			for _, s := range tt.contains {
				if !strings.Contains(resultStr, s) {
					t.Errorf("Expected result to contain %q, got %q", s, resultStr)
				}
			}
		})
	}
}
