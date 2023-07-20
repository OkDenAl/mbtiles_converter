package utils

import (
	"math"
)

// Lon2tileFloor converts longitude to tile x coordinate and applies the operation math.Floor()
func Lon2tileFloor(lon float64, zoom int) float64 {
	return math.Floor(Lon2tile(lon, zoom))
}

// Lon2tile converts longitude to tile x coordinate
func Lon2tile(lon float64, zoom int) float64 {
	return (lon + 180.0) / 360 * math.Pow(2, float64(zoom))
}

// Lat2tileFloor converts latitude to tile y coordinate and applies the operation math.Floor()
func Lat2tileFloor(lat float64, zoom int) float64 {
	return math.Floor(Lat2tile(lat, zoom))
}

// Lat2tile converts latitude to tile y coordinate
func Lat2tile(lat float64, zoom int) float64 {
	return (1 - math.Log(math.Tan(lat*math.Pi/180)+1/math.Cos(lat*math.Pi/180))/math.Pi) / 2 * math.Pow(2, float64(zoom))
}
