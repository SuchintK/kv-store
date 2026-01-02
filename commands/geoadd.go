package command

import (
	"fmt"
	"strconv"

	"github.com/SuchintK/GoDisKV/geohash"
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type GeoAddCommand Command

const (
	minLatitude  = geohash.MinLatitude
	maxLatitude  = geohash.MaxLatitude
	minLongitude = geohash.MinLongitude
	maxLongitude = geohash.MaxLongitude
)

func (cmd *GeoAddCommand) Execute(con *client.Client) RESPValue {
	// GEOADD key longitude latitude member [longitude latitude member ...]
	// Minimum: key + lon + lat + member = 4 args
	if len(cmd.args) < 4 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	// Check if arguments come in groups of 3 (longitude, latitude, member)
	if (len(cmd.args)-1)%3 != 0 {
		return resp.EncodeSimpleError("ERR syntax error")
	}

	key := cmd.args[0]
	addedCount := 0

	// Get or create sorted set
	val, exists := store.Get(key)
	var sortedSet *store.SortedSet

	if !exists {
		sortedSet = store.NewSortedSet()
		store.Set(key, &store.Value{SortedSetData: sortedSet})
	} else {
		if val.SortedSetData == nil {
			return resp.EncodeSimpleError(errWrongType)
		}
		sortedSet = val.SortedSetData
	}

	// Process each location (groups of 3: longitude, latitude, member)
	for i := 1; i < len(cmd.args); i += 3 {
		lonStr := cmd.args[i]
		latStr := cmd.args[i+1]
		member := cmd.args[i+2]

		// Parse longitude
		longitude, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			return resp.EncodeSimpleError("ERR value is not a valid float")
		}

		// Parse latitude
		latitude, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			return resp.EncodeSimpleError("ERR value is not a valid float")
		}

		// Validate longitude range
		if longitude < minLongitude || longitude > maxLongitude {
			return resp.EncodeSimpleError(fmt.Sprintf("ERR invalid longitude, must be between %.10f and %.10f", minLongitude, maxLongitude))
		}

		// Validate latitude range
		if latitude < minLatitude || latitude > maxLatitude {
			return resp.EncodeSimpleError(fmt.Sprintf("ERR invalid latitude, must be between %.10f and %.10f", minLatitude, maxLatitude))
		}

		// Encode coordinates to geohash score
		geohashScore := geohash.Encode(latitude, longitude)

		// Add to sorted set
		// Returns true if member was newly added, false if score was updated
		if sortedSet.Add(float64(geohashScore), member) {
			addedCount++
		}
	}

	return resp.EncodeInteger(int64(addedCount))
}
