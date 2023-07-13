package entity

type MapPoint struct {
	Longitude float64
	Latitude  float64
	Type      string
}

type MbtilesMapPoint struct {
	TileRow   float64
	TileCol   float64
	ZoomLevel int
	TileData  []byte
}
