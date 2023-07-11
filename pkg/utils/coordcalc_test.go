package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type tCase struct {
	name   string
	lon    float64
	lat    float64
	zoom   int
	expLon float64
	expLat float64
}

func TestCoordCalc(t *testing.T) {
	tests := []tCase{
		{
			name: "Moscow, zoom = 6", lon: 37.6156, lat: 55.7522, zoom: 6,
			expLon: float64(38), expLat: float64(20),
		},
		{
			name: "Moscow, zoom = 18", lon: 37.6156, lat: 55.7522, zoom: 18,
			expLon: float64(158462), expLat: float64(81951),
		},
		{
			name: "Australia city, zoom = 6", lon: 144.963, lat: -37.814, zoom: 6,
			expLon: float64(57), expLat: float64(39),
		},
	}
	for _, tc := range tests {
		assert.Equal(t, Lon2tileFloor(tc.lon, tc.zoom), tc.expLon)
		assert.Equal(t, Lat2tileFloor(tc.lat, tc.zoom), tc.expLat)
	}
}
