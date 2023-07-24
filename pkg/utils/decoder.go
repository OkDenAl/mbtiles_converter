package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/pkg/mvt"
	"io"
)

// DecodeFromGzipMVT decodes tile_data from .mvt.gz format to *mvt.Tile
func DecodeFromGzipMVT(data []byte) (*mvt.Tile, error) {
	rdata := bytes.NewReader(data)
	gzreader, err := gzip.NewReader(rdata)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader: %w", err)
	}
	output, err := io.ReadAll(gzreader)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}
	tile, err := mvt.DecodeByte(output)
	if err != nil {
		return nil, fmt.Errorf("mvt.DecodeByte: %w", err)
	}
	return tile, nil
}
