package command

import (
	"sort"
	"strconv"
	"strings"

	"github.com/SuchintK/GoDisKV/geohash"
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type GeoRadiusCommand Command

type geoRadiusResult struct {
	member   string
	distance float64
	lat      float64
	lon      float64
}

func (cmd *GeoRadiusCommand) Execute(con *client.Client) RESPValue {
	// GEORADIUS key longitude latitude radius m|km|ft|mi [WITHCOORD] [WITHDIST] [WITHHASH] [COUNT count] [ASC|DESC]
	if len(cmd.args) < 5 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]

	// Parse longitude
	longitude, err := strconv.ParseFloat(cmd.args[1], 64)
	if err != nil {
		return resp.EncodeSimpleError("ERR value is not a valid float")
	}

	// Parse latitude
	latitude, err := strconv.ParseFloat(cmd.args[2], 64)
	if err != nil {
		return resp.EncodeSimpleError("ERR value is not a valid float")
	}

	// Parse radius
	radius, err := strconv.ParseFloat(cmd.args[3], 64)
	if err != nil {
		return resp.EncodeSimpleError("ERR value is not a valid float")
	}

	// Parse unit
	unit := strings.ToLower(cmd.args[4])
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

	// Convert radius to meters
	radiusMeters := radius / conversionFactor

	// Parse optional arguments
	withCoord := false
	withDist := false
	withHash := false
	count := -1
	ascending := false
	descending := false

	for i := 5; i < len(cmd.args); i++ {
		arg := strings.ToUpper(cmd.args[i])
		switch arg {
		case "WITHCOORD":
			withCoord = true
		case "WITHDIST":
			withDist = true
		case "WITHHASH":
			withHash = true
		case "COUNT":
			if i+1 >= len(cmd.args) {
				return resp.EncodeSimpleError("ERR syntax error")
			}
			i++
			var err error
			count, err = strconv.Atoi(cmd.args[i])
			if err != nil || count < 0 {
				return resp.EncodeSimpleError("ERR value is out of range, must be positive")
			}
		case "ASC":
			ascending = true
		case "DESC":
			descending = true
		default:
			return resp.EncodeSimpleError("ERR syntax error")
		}
	}

	// Get sorted set
	val, exists := store.Get(key)
	if !exists {
		return resp.EncodeArray([][]byte{})
	}

	if val.SortedSetData == nil {
		return resp.EncodeSimpleError(errWrongType)
	}

	sortedSet := val.SortedSetData

	// Find all members within radius
	results := []geoRadiusResult{}

	// Get all members from sorted set
	allMembers := sortedSet.GetRange(0, sortedSet.Card()-1)

	for _, member := range allMembers {
		score, _ := sortedSet.GetScore(member)
		memberLat, memberLon := geohash.Decode(uint64(score))

		// Calculate distance
		distanceMeters := geohash.Distance(latitude, longitude, memberLat, memberLon)

		// Check if within radius
		if distanceMeters <= radiusMeters {
			results = append(results, geoRadiusResult{
				member:   member,
				distance: distanceMeters * conversionFactor,
				lat:      memberLat,
				lon:      memberLon,
			})
		}
	}

	// Sort results
	if ascending {
		sort.Slice(results, func(i, j int) bool {
			return results[i].distance < results[j].distance
		})
	} else if descending {
		sort.Slice(results, func(i, j int) bool {
			return results[i].distance > results[j].distance
		})
	}

	// Apply count limit
	if count > 0 && count < len(results) {
		results = results[:count]
	}

	// Encode results
	respResults := make([][]byte, len(results))
	for i, result := range results {
		if !withCoord && !withDist && !withHash {
			// Simple member name only
			respResults[i] = resp.EncodeBulkString(result.member)
		} else {
			// Build array with member and optional fields
			fields := [][]byte{resp.EncodeBulkString(result.member)}

			if withDist {
				fields = append(fields, resp.EncodeBulkString(formatFloat(result.distance)))
			}

			if withHash {
				hash := geohash.Encode(result.lat, result.lon)
				fields = append(fields, resp.EncodeInteger(int64(hash)))
			}

			if withCoord {
				coords := [][]byte{
					resp.EncodeBulkString(formatFloat(result.lon)),
					resp.EncodeBulkString(formatFloat(result.lat)),
				}
				fields = append(fields, resp.EncodeArray(coords))
			}

			respResults[i] = resp.EncodeArray(fields)
		}
	}

	return resp.EncodeArray(respResults)
}
