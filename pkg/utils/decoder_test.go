package utils

import (
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	mvt2 "github.com/OkDenAl/mbtiles_converter/pkg/mvt"
	"github.com/go-spatial/geom"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestDecoder(t *testing.T) {
	mvt, err := EncodePixelCoordToGzipMVT([]entity.TilePoint{{10, 9, "cafe"}, {11, 9, "bebra"}}, 6)
	log.Println(mvt)
	assert.NoError(t, err)
	ti := time.Now()
	a, err := DecodeFromGzipMVT(mvt)
	log.Println(*a)
	assert.NoError(t, err)
	assert.Equal(t, a.TakeLayers()[0].Version(), 2)
	assert.Equal(t, a.TakeLayers()[0].Extent(), 64)
	f := mvt2.NewFeatures(geom.Point{11, 9}, nil)
	a.Layers[0] = a.TakeLayers()[0].AddFeatures(f...)
	toMVT, err := EncodeTileToMVT(*a)
	assert.NoError(t, err)
	log.Println(time.Since(ti))
	log.Println(toMVT)
}
