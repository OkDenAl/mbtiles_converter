package utils

import (
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecoder(t *testing.T) {
	mvt, err := EncodePixelCoordToGzipMVT([]entity.TilePoint{{X: 10, Y: 9, Type: "cafe"}}, 6)
	assert.NoError(t, err)
	a, err := DecodeFromGzipMVT(mvt)
	assert.NoError(t, err)
	assert.Equal(t, a.TakeLayers()[0].Version(), 2)
	assert.Equal(t, a.TakeLayers()[0].Extent(), 64)
	assert.Len(t, a.TakeLayers(), 1)
	assert.Len(t, a.TakeLayers()[0].Features, 1)
}
