package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecoder(t *testing.T) {
	mvt, err := EncodePixelCoordToGzipMVT([][2]float64{{10, 8}}, 6)
	assert.NoError(t, err)
	a, err := DecodeFromGzipMVT(mvt)
	assert.NoError(t, err)
	assert.Equal(t, a.Layers()[0].Version(), 2)
	assert.Equal(t, a.Layers()[0].Extent(), 64)
}
