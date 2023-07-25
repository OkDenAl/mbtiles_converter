package utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"github.com/OkDenAl/mbtiles_converter/pkg/mvt"
	"github.com/go-spatial/geom"
	"google.golang.org/protobuf/proto"
)

// EncodePixelCoordToGzipMVT encodes tile points to mvt format
func EncodePixelCoordToGzipMVT(tilePoints []entity.TilePoint, zoom int) ([]byte, error) {
	l := &mvt.Layer{Name: entity.DefaultLayerName}
	features := make([]mvt.Feature, len(tilePoints))
	for i, tilePoint := range tilePoints {
		geo := geom.Point{tilePoint.X, tilePoint.Y}
		features[i] = mvt.Feature{Geometry: geo, Tags: tilePoint.Tags}
	}
	l.AddFeatures(features...)
	l.SetExtent(1 << zoom)
	t := mvt.Tile{}
	err := t.AddLayers(l)
	if err != nil {
		return nil, fmt.Errorf("t.AddLayers: %w", err)
	}
	mvtTile, err := EncodeTileToMVT(t)
	if err != nil {
		return nil, fmt.Errorf("EncodeTileToMVT: %w", err)
	}
	return mvtTile, nil
}

// EncodeTileToMVT encodes tile to mvt format
func EncodeTileToMVT(t mvt.Tile) ([]byte, error) {
	tile, err := t.VTile(context.Background())
	if err != nil {
		return nil, fmt.Errorf("t.VTile: %w", err)
	}
	marshal, err := proto.Marshal(tile)
	if err != nil {
		return nil, fmt.Errorf("proto.Marshal: %w", err)
	}

	var gzipBuf bytes.Buffer
	w := gzip.NewWriter(&gzipBuf)
	_, err = w.Write(marshal)
	if err != nil {
		return nil, fmt.Errorf("w.Write: %w", err)
	}
	_ = w.Close()
	return gzipBuf.Bytes(), nil
}
