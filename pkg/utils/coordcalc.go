package utils

import (
	"math"
)

//function lon2tile(lon,zoom) { return (Math.floor((lon+180)/360*Math.pow(2,zoom))); }
//function lat2tile(lat,zoom)  { return (Math.floor((1-Math.log(Math.tan(lat*Math.PI/180) + 1/Math.cos(lat*Math.PI/180))/Math.PI)/2 *Math.pow(2,zoom))); }

func Lon2tileFloor(lon float64, zoom int) float64 {
	return math.Floor(Lon2tile(lon, zoom))
}

func Lon2tile(lon float64, zoom int) float64 {
	return (lon + 180.0) / 360 * math.Pow(2, float64(zoom))
}

func Lat2tileFloor(lat float64, zoom int) float64 {
	return math.Floor(Lat2tile(lat, zoom))
}

func Lat2tile(lat float64, zoom int) float64 {
	return (1 - math.Log(math.Tan(lat*math.Pi/180)+1/math.Cos(lat*math.Pi/180))/math.Pi) / 2 * math.Pow(2, float64(zoom))
}
