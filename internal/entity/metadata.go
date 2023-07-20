package entity

import (
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/config"
)

const DefaultLayerName = "objects"

// Metadata represented the metadata table rows
type Metadata struct {
	Name    string
	Format  string
	Bounds  string
	Center  string
	Type    string
	MinZoom int
	MaxZoom int
	Json    string
}

// NewMetadata creates metadata from config data
func NewMetadata(meta config.Metadata) Metadata {
	return Metadata{Name: meta.Name, Bounds: meta.Bounds, Center: meta.Center, Type: "overlay",
		MinZoom: meta.MinZoom, MaxZoom: meta.MaxZoom, Format: "pbf", Json: makeJson(meta)}
}

func makeJson(meta config.Metadata) string {
	return fmt.Sprintf("{\"vector_layers\": [ { \"id\": \"%s\", \"description\": \"\", \"minzoom\": %d, \"maxzoom\": %d } ]}",
		DefaultLayerName, meta.MinZoom, meta.MaxZoom)
}
