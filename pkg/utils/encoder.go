package utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"github.com/OkDenAl/mbtiles_converter/pkg/mvt"
	"github.com/go-spatial/geom"
	"google.golang.org/protobuf/proto"
)

func EncodePixelCoordToGzipMVT(tilePoints []entity.TilePoint, zoom int) ([]byte, error) {
	l := &mvt.Layer{Name: entity.DefaultLayerName}
	features := make([]mvt.Feature, len(tilePoints))
	for i, tilePoint := range tilePoints {
		geo := geom.Point{tilePoint.X, tilePoint.Y}
		features[i] = mvt.Feature{Geometry: geo, Tags: map[string]interface{}{"type": tilePoint.Type}}
	}
	l.AddFeatures(features...)
	l.SetExtent(1 << zoom)
	t := mvt.Tile{}
	err := t.AddLayers(l)
	if err != nil {
		return nil, err
	}
	return EncodeTileToMVT(t)
}

func EncodeTileToMVT(t mvt.Tile) ([]byte, error) {
	tile, err := t.VTile(context.Background())
	if err != nil {
		return nil, err
	}
	marshal, err := proto.Marshal(tile)
	if err != nil {
		return nil, err
	}

	var gzipBuf bytes.Buffer
	w := gzip.NewWriter(&gzipBuf)
	_, err = w.Write(marshal)
	if err != nil {
		return nil, err
	}
	_ = w.Close()
	return gzipBuf.Bytes(), nil
}
