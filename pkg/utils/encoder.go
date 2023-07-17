package utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/OkDenAl/mbtiles_converter/pkg/mvt"
	"github.com/go-spatial/geom"
	"github.com/golang/protobuf/proto"
)

const DefaultLayerName = "cities"

func EncodePixelCoordToGzipMVT(coords [][2]float64, zoom int) ([]byte, error) {
	geo := geom.MultiPoint{}
	err := geo.SetPoints(coords)
	if err != nil {
		return nil, err
	}
	f := mvt.NewFeatures(geo, nil)
	l := &mvt.Layer{Name: DefaultLayerName}
	l.AddFeatures(f...)
	l.SetExtent(1 << zoom)
	t := mvt.Tile{}
	err = t.AddLayers(l)
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
