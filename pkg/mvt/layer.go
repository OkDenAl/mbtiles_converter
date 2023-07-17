package mvt

import (
	"fmt"
	vectorTile "github.com/OkDenAl/mbtiles_converter/pkg/mvt/vector_tile"

	"context"
)

const (
	Version uint32 = 2
)

// Layer describes a layer within a tile.
// Each layer can have multiple features
type Layer struct {
	// Name is the unique name of the layer within the tile
	Name string
	// The set of features
	features []Feature
	// default is 4096
	extent *int
}

func valMapToVTileValue(valMap []interface{}) (vt []*vectorTile.Tile_Value) {
	for _, v := range valMap {
		vt = append(vt, vectorTileValue(v))
	}

	return vt
}

// VTileLayer returns a vectorTile Tile_Layer object that represents this layer.
func (l *Layer) VTileLayer(ctx context.Context) (*vectorTile.Tile_Layer, error) {
	kmap, vmap, err := keyvalMapsFromFeatures(l.features)
	if err != nil {
		return nil, err
	}

	valmap := valMapToVTileValue(vmap)

	var features = make([]*vectorTile.Tile_Feature, 0, len(l.features))
	for _, f := range l.features {
		vtf, err := f.VTileFeature(ctx, kmap, vmap)
		if err != nil {
			switch err {
			case context.Canceled:
				return nil, err
			default:
				return nil, fmt.Errorf("error getting VTileFeature: %v", err)
			}
		}

		if vtf != nil {
			features = append(features, vtf)
		}
	}
	ext := uint32(*l.extent)
	version := Version
	vtl := new(vectorTile.Tile_Layer)
	vtl.Version = version
	name := l.Name
	vtl.Name = name
	vtl.Features = features
	vtl.Keys = kmap
	vtl.Values = valmap
	vtl.Extent = ext
	return vtl, nil
}

// SetExtent sets the extent value
func (l *Layer) SetExtent(e int) {
	if l == nil {
		return
	}
	l.extent = &e
}

// Features returns a copy of the features in the layer, use the index of the this
// array to remove any features from the layer
func (l *Layer) Features() (f []Feature) {
	if l == nil || l.features == nil {
		return nil
	}
	f = append(f, l.features...)
	return f
}

// AddFeatures will add one or more Features to the Layer
// per the spec features SHOULD have unique ids but it's not required
func (l *Layer) AddFeatures(features ...Feature) Layer {
	// pre allocate memory
	b := make([]Feature, len(l.features)+len(features))

	copy(b, l.features)
	copy(b[len(l.features):], features)

	l.features = b
	return *l
}

func vectorTileValue(i interface{}) *vectorTile.Tile_Value {
	tv := new(vectorTile.Tile_Value)
	switch t := i.(type) {

	case int:
		tv.Type = &vectorTile.Tile_Value_IntValue{IntValue: int64(t)}
	case string:
		tv.Type = &vectorTile.Tile_Value_StringValue{StringValue: t}
	default:
		return nil
	}
	return tv
}
