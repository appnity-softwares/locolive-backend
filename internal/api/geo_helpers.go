package api

import "github.com/mmcloughlin/geohash"

// truncatedGeohash generates a geohash and truncates it to the specified precision
func truncatedGeohash(lat, lng float64, precision int) string {
	hash := geohash.Encode(lat, lng)
	if len(hash) > precision {
		return hash[:precision]
	}
	return hash
}
