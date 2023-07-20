package entity

// MapPoint represented the data in PostgreSQL table (you can change this for your database BUT
// Longitude and Latitude always must come first)
type MapPoint struct {
	Longitude float64
	Latitude  float64
	Type      string
}

// MbtilesMapPoint represented the data in SQLite table
type MbtilesMapPoint struct {
	TileRow   float64
	TileCol   float64
	ZoomLevel int
	TileData  []byte
}

// TilePoint represented the pixel coordinates in tile
type TilePoint struct {
	X    float64
	Y    float64
	Type string
}
