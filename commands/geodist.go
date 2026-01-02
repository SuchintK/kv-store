package command

import (
	"strconv"
	"strings"

	"github.com/SuchintK/GoDisKV/geohash"
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type GeoDistCommand Command

// Unit conversion factors (from meters)
const (
	metersToMeters     = 1.0
	metersToKilometers = 0.001
	metersToMiles      = 0.000621371
	metersToFeet       = 3.28084
)

func (cmd *GeoDistCommand) Execute(con *client.Client) RESPValue {
	// GEODIST key member1 member2 [unit]
	if len(cmd.args) < 3 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]
	member1 := cmd.args[1]
	member2 := cmd.args[2]

	// Default unit is meters
	unit := "m"
	if len(cmd.args) >= 4 {
		unit = strings.ToLower(cmd.args[3])
	}

	// Validate unit
	var conversionFactor float64
	switch unit {
	case "m":
		conversionFactor = metersToMeters
	case "km":
		conversionFactor = metersToKilometers
	case "mi":
		conversionFactor = metersToMiles
	case "ft":
		conversionFactor = metersToFeet
	default:
		return resp.EncodeSimpleError("ERR unsupported unit provided. please use M, KM, FT, MI")
	}

	// Get sorted set
	val, exists := store.Get(key)
	if !exists {
		return resp.EncodeNullBulkString()
	}

	if val.SortedSetData == nil {
		return resp.EncodeSimpleError(errWrongType)
	}

	sortedSet := val.SortedSetData

	// Get scores for both members
	score1, exists1 := sortedSet.GetScore(member1)
	if !exists1 {
		return resp.EncodeNullBulkString()
	}

	score2, exists2 := sortedSet.GetScore(member2)
	if !exists2 {
		return resp.EncodeNullBulkString()
	}

	// Decode geohashes to get coordinates
	lat1, lon1 := geohash.Decode(uint64(score1))
	lat2, lon2 := geohash.Decode(uint64(score2))

	// Calculate distance in meters
	distanceMeters := geohash.Distance(lat1, lon1, lat2, lon2)

	// Convert to requested unit
	distance := distanceMeters * conversionFactor

	return resp.EncodeBulkString(formatFloat(distance))
}

// formatFloat formats a float64 for RESP output with appropriate precision
func formatFloat(f float64) string {
	// Use 'f' format with enough precision, then trim trailing zeros
	s := strconv.FormatFloat(f, 'f', 10, 64)
	// Remove trailing zeros and unnecessary decimal point
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
