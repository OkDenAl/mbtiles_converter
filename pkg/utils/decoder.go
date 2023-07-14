package utils

import (
	"bytes"
	"compress/gzip"
	"github.com/go-spatial/geom/encoding/mvt"
	"io"
)

func DecodeFromGzipMVT(data []byte) (*mvt.Tile, error) {
	rdata := bytes.NewReader(data)
	gzreader, err := gzip.NewReader(rdata)
	if err != nil {
		return nil, err
	}
	output, err := io.ReadAll(gzreader)
	if err != nil {
		return nil, err
	}
	return mvt.DecodeByte(output)
}