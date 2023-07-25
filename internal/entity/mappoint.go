package entity

// MapPoint represented the data in PostgreSQL table
type MapPoint struct {
	Longitude      float64
	Latitude       float64
	AdditionalRows map[string]any
}

type AdditionalRows struct {
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
	Tags map[string]any
}
