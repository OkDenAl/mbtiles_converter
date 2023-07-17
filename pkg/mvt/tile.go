package mvt

import (
	"context"
	"fmt"
	vectorTile "github.com/OkDenAl/mbtiles_converter/pkg/mvt/vector_tile"
)

// Tile describes a Mapbox Vector Tile
type Tile struct {
	Layers []Layer
}

// AddLayers adds a Layer to the Tile
func (t *Tile) AddLayers(layers ...*Layer) error {
	for i := range layers {
		nl := layers[i]
		if nl == nil {
			continue
		}
		for i, l := range t.Layers {
			if l.Name == nl.Name {
				return fmt.Errorf("layer %v, already is named %v, new layer not added.", i, l.Name)
			}
		}
		t.Layers = append(t.Layers, *nl)
	}
	return nil
}

// Layers returns a copy of the layers in this tile.
func (t *Tile) TakeLayers() (l []Layer) {
	l = append(l, t.Layers...)
	return l
}

// Version is the version of tile spec this layer is from.
func (*Layer) Version() int { return int(Version) }

// Extent defaults to 4096
func (l *Layer) Extent() int {
	if l == nil || l.extent == nil {
		return int(64)
	}
	return *(l.extent)
}

// VTile returns a Tile according to the Google Protobuff definition.
// This function does the hard work of converting everything to the standard.
func (t *Tile) VTile(ctx context.Context) (vt *vectorTile.Tile, err error) {
	vt = new(vectorTile.Tile)

	for _, l := range t.Layers {
		vtl, err := l.VTileLayer(ctx)
		if err != nil {
			switch err {
			case context.Canceled:
				return nil, err
			default:
				return nil, fmt.Errorf("error Getting VTileLayer: %v", err)
			}
		}

		vt.Layers = append(vt.Layers, vtl)
	}

	return vt, nil
}
