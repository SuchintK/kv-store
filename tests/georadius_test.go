package tests

import (
	"net"
	"strings"
	"testing"

	command "github.com/SuchintK/GoDisKV/commands"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

func setupGeoRadiusTestClient() *client.Client {
	conn1, _ := net.Pipe()
	cli := client.New(conn1)
	return &cli
}

func TestGeoRadiusCommand(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		args        []string
		contains    []string
		notContains []string
	}{
		{
			name: "Find locations within radius",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoRadiusTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo", "15.087269", "37.502669", "Catania"}).Execute(cli)
			},
			args:     []string{"locations", "15.0", "37.0", "200", "km"},
			contains: []string{"Palermo", "Catania"},
		},
		{
			name: "Find locations with smaller radius",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoRadiusTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo", "15.087269", "37.502669", "Catania"}).Execute(cli)
			},
			args:        []string{"locations", "15.0", "37.0", "100", "km"},
			contains:    []string{"Catania"},
			notContains: []string{"Palermo"},
		},
		{
			name: "GEORADIUS with WITHDIST",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoRadiusTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo"}).Execute(cli)
			},
			args:     []string{"locations", "15.0", "37.0", "200", "km", "WITHDIST"},
			contains: []string{"Palermo"},
		},
		{
			name: "GEORADIUS with WITHCOORD",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoRadiusTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo"}).Execute(cli)
			},
			args:     []string{"locations", "15.0", "37.0", "200", "km", "WITHCOORD"},
			contains: []string{"Palermo", "13.361"}, // Check prefix due to encoding precision
		},
		{
			name: "GEORADIUS with COUNT",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoRadiusTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo", "15.087269", "37.502669", "Catania", "12.5", "37.8", "Agrigento"}).Execute(cli)
			},
			args:     []string{"locations", "15.0", "37.0", "300", "km", "COUNT", "2"},
			contains: []string{"*2\r\n"},
		},
		{
			name: "Empty result when no locations in radius",
			setup: func() {
				store.Delete("locations")
				cli := setupGeoRadiusTestClient()
				command.New("geoadd", []string{"locations", "13.361389", "38.115556", "Palermo"}).Execute(cli)
			},
			args:     []string{"locations", "0.0", "0.0", "100", "km"},
			contains: []string{"*0\r\n"},
		},
		{
			name: "Non-existent key returns empty array",
			setup: func() {
				store.Delete("nonexistent")
			},
			args:     []string{"nonexistent", "15.0", "37.0", "100", "km"},
			contains: []string{"*0\r\n"},
		},
		{
			name:     "Error on wrong number of arguments",
			setup:    func() {},
			args:     []string{"locations", "15.0", "37.0"},
			contains: []string{"-wrong number of arguments"},
		},
		{
			name:     "Error on invalid unit",
			setup:    func() {},
			args:     []string{"locations", "15.0", "37.0", "100", "invalid"},
			contains: []string{"-ERR unsupported unit"},
		},
		{
			name: "Error on wrong type",
			setup: func() {
				store.Delete("mystring")
				cli := setupGeoRadiusTestClient()
				command.New("set", []string{"mystring", "value"}).Execute(cli)
			},
			args:     []string{"mystring", "15.0", "37.0", "100", "km"},
			contains: []string{"-WRONGTYPE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cli := setupGeoRadiusTestClient()
			cmd := command.New("georadius", tt.args)
			result := cmd.Execute(cli)
			resultStr := string(result)

			for _, s := range tt.contains {
				if !strings.Contains(resultStr, s) {
					t.Errorf("Expected result to contain %q, got %q", s, resultStr)
				}
			}

			for _, s := range tt.notContains {
				if strings.Contains(resultStr, s) {
					t.Errorf("Expected result NOT to contain %q, got %q", s, resultStr)
				}
			}
		})
	}
}
