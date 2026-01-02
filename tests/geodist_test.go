package tests

import (
	"net"
	"strings"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupGeoDistTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestGeoDistCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		args     []string
		contains string
	}{
		{
			name: "Calculate distance between two cities in meters",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoDistTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo", "15.087269", "37.502669", "Catania"}).Execute(cli)
			},
			args:     []string{"locations", "Palermo", "Catania"},
			contains: "166",
		},
		{
			name: "Calculate distance in kilometers",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoDistTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo", "15.087269", "37.502669", "Catania"}).Execute(cli)
			},
			args:     []string{"locations", "Palermo", "Catania", "km"},
			contains: "166",
		},
		{
			name: "Calculate distance in miles",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoDistTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo", "15.087269", "37.502669", "Catania"}).Execute(cli)
			},
			args:     []string{"locations", "Palermo", "Catania", "mi"},
			contains: "103",
		},
		{
			name: "Distance from member to itself is zero",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoDistTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo"}).Execute(cli)
			},
			args:     []string{"locations", "Palermo", "Palermo"},
			contains: "0",
		},
		{
			name: "Non-existent first member returns null",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoDistTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo"}).Execute(cli)
			},
			args:     []string{"locations", "NonExistent", "Palermo"},
			contains: "$-1",
		},
		{
			name: "Non-existent key returns null",
			setup: func() {
				store.Delete("nonexistent")
			},
			args:     []string{"nonexistent", "member1", "member2"},
			contains: "$-1",
		},
		{
			name: "Error on invalid unit",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoDistTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo", "15.087269", "37.502669", "Catania"}).Execute(cli)
			},
			args:     []string{"locations", "Palermo", "Catania", "invalid"},
			contains: "-ERR unsupported unit",
		},
		{
			name:     "Error on wrong number of arguments",
			setup:    func() {},
			args:     []string{"locations", "Palermo"},
			contains: "-wrong number of arguments",
		},
		{
			name: "Error on wrong type",
			setup: func() {
				store.Delete("mystring")
				cli := setupGeoDistTestClient()
				command.New("set", []string{"mystring", "value"}).Execute(cli)
			},
			args:     []string{"mystring", "member1", "member2"},
			contains: "-WRONGTYPE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupGeoDistTestClient()
			cmd := command.New("geodist", tt.args)
			result := cmd.Execute(cli)
			resultStr := string(result)

			if tt.contains != "" {
				if !strings.Contains(resultStr, tt.contains) {
					t.Errorf("Expected result to contain %q, got %q", tt.contains, resultStr)
				}
			}
		})
	}
}
