package geohash

import "math"

const (
	MinLatitude  = -85.05112878
	MaxLatitude  = 85.05112878
	MinLongitude = -180.0
	MaxLongitude = 180.0

	latitudeRange  = MaxLatitude - MinLatitude
	longitudeRange = MaxLongitude - MinLongitude

	// Earth radius in meters
	EarthRadiusMeters = 6372797.560856
)

func spreadInt32ToInt64(v uint32) uint64 {
	result := uint64(v)
	result = (result | (result << 16)) & 0x0000FFFF0000FFFF
	result = (result | (result << 8)) & 0x00FF00FF00FF00FF
	result = (result | (result << 4)) & 0x0F0F0F0F0F0F0F0F
	result = (result | (result << 2)) & 0x3333333333333333
	result = (result | (result << 1)) & 0x5555555555555555
	return result
}

func interleave(x, y uint32) uint64 {
	xSpread := spreadInt32ToInt64(x)
	ySpread := spreadInt32ToInt64(y)
	yShifted := ySpread << 1
	return xSpread | yShifted
}

// Encode converts latitude and longitude to a 52-bit geohash score
func Encode(latitude, longitude float64) uint64 {
	// Normalize to the range 0-2^26
	normalizedLatitude := math.Pow(2, 26) * (latitude - MinLatitude) / latitudeRange
	normalizedLongitude := math.Pow(2, 26) * (longitude - MinLongitude) / longitudeRange

	// Truncate to integers
	latInt := uint32(normalizedLatitude)
	lonInt := uint32(normalizedLongitude)

	return interleave(latInt, lonInt)
}

func unspreadInt64ToInt32(v uint64) uint32 {
	result := v & 0x5555555555555555
	result = (result | (result >> 1)) & 0x3333333333333333
	result = (result | (result >> 2)) & 0x0F0F0F0F0F0F0F0F
	result = (result | (result >> 4)) & 0x00FF00FF00FF00FF
	result = (result | (result >> 8)) & 0x0000FFFF0000FFFF
	result = (result | (result >> 16)) & 0x00000000FFFFFFFF
	return uint32(result)
}

func deinterleave(hash uint64) (uint32, uint32) {
	x := unspreadInt64ToInt32(hash)
	y := unspreadInt64ToInt32(hash >> 1)
	return x, y
}

// Decode converts a 52-bit geohash score back to latitude and longitude
func Decode(hash uint64) (latitude, longitude float64) {
	latInt, lonInt := deinterleave(hash)

	// Denormalize back to original range
	latitude = float64(latInt)*latitudeRange/math.Pow(2, 26) + MinLatitude
	longitude = float64(lonInt)*longitudeRange/math.Pow(2, 26) + MinLongitude

	return latitude, longitude
}

// Distance calculates the distance between two points using the Haversine formula
// Returns distance in meters
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusMeters * c
}
