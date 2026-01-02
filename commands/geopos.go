package command

import (
	"github.com/SuchintK/GoDisKV/geohash"
	"github.com/SuchintK/GoDisKV/resp"
	"github.com/SuchintK/GoDisKV/resp/client"
	"github.com/SuchintK/GoDisKV/store"
)

type GeoPosCommand Command

func (cmd *GeoPosCommand) Execute(con *client.Client) RESPValue {
	// GEOPOS key member [member ...]
	if len(cmd.args) < 2 {
		return resp.EncodeSimpleError(errWrongNumberOfArgs)
	}

	key := cmd.args[0]
	members := cmd.args[1:]

	// Get sorted set
	val, exists := store.Get(key)
	if !exists {
		// Return array of nils for non-existent key
		results := make([][]byte, len(members))
		for i := range results {
			results[i] = resp.EncodeNullBulkString()
		}
		return resp.EncodeArray(results)
	}

	if val.SortedSetData == nil {
		return resp.EncodeSimpleError(errWrongType)
	}

	sortedSet := val.SortedSetData
	results := make([][]byte, len(members))

	for i, member := range members {
		score, exists := sortedSet.GetScore(member)
		if !exists {
			results[i] = resp.EncodeNullBulkString()
			continue
		}

		// Decode geohash to get latitude and longitude
		lat, lon := geohash.Decode(uint64(score))

		// Return array of [longitude, latitude]
		coords := [][]byte{
			resp.EncodeBulkString(formatFloat(lon)),
			resp.EncodeBulkString(formatFloat(lat)),
		}
		results[i] = resp.EncodeArray(coords)
	}

	return resp.EncodeArray(results)
}
