package mvt

import (
	"context"
	"fmt"
	vectorTile "github.com/OkDenAl/mbtiles_converter/pkg/mvt/vector_tile"
	"log"

	"github.com/go-spatial/geom"
)

var (
	ErrNilFeature          = fmt.Errorf("feature is nil")
	ErrUnknownGeometryType = fmt.Errorf("unknown geometry type")
	ErrNilGeometryType     = fmt.Errorf("geometry is nil")
)

type Feature struct {
	ID       *uint64
	Tags     map[string]interface{}
	Geometry geom.Geometry
}

// NewFeatures returns one or more features for the given Geometry
func NewFeatures(geo geom.Geometry, tags map[string]interface{}) (f []Feature) {
	if geo == nil {
		return f // return empty feature set for a nil geometry
	}

	if g, ok := geo.(geom.Collection); ok {
		geos := g.Geometries()
		for i := range geos {
			f = append(f, NewFeatures(geos[i], tags)...)
		}
		return f
	}
	f = append(f, Feature{
		Tags:     tags,
		Geometry: geo,
	})
	return f
}

// VTileFeature will return a vectorTile.Feature that would represent the Feature
func (f *Feature) VTileFeature(ctx context.Context, keys []string, vals []interface{}) (tf *vectorTile.Tile_Feature, err error) {
	tf = new(vectorTile.Tile_Feature)
	tf.Id = 1

	if tf.Tags, err = keyvalTagsMap(keys, vals, f); err != nil {
		return tf, err
	}

	geo, gtype, err := encodeGeometry(ctx, f.Geometry)
	if err != nil {
		return tf, err
	}

	if len(geo) == 0 {
		return nil, nil
	}

	tf.Geometry = geo
	tf.Type = gtype

	return tf, nil
}

const (
	cmdMoveTo uint32 = 1
)

type Command uint32

// NewCommand return a new command encoder
func NewCommand(cmd uint32, count int) Command {
	return Command((cmd & 0x7) | (uint32(count) << 3))
}

// encodeZigZag does the ZigZag encoding for small ints.
func encodeZigZag(i int64) uint32 {
	return uint32((i << 1) ^ (i >> 31))
}

// ID encodes the ID of the command
func (c Command) ID() uint32 {
	return uint32(c) & 0x7
}

// Count encode the count of elements in the command
func (c Command) Count() int {
	return int(uint32(c) >> 3)
}

// cursor reprsents the current position, this is needed to encode the geometry.
// the origin (0,0) is the top-left of the Tile.
type cursor struct {
	x int64
	y int64
}

// NewCursor creates a new cursor for drawing and MVT tile
func NewCursor() *cursor {
	return &cursor{}
}

// GetDeltaPointAndUpdate returns the delta of for the given point from the current
// cursor position
func (c *cursor) GetDeltaPointAndUpdate(p geom.Point) (dx, dy int64) {
	delta := c.moveCursorPoints([2]int64{int64(p.X()), int64(p.Y())})
	return delta[0][0], delta[0][1]
}

func (c *cursor) moveCursorPoints(pts ...[2]int64) (deltas [][2]int64) {
	deltas = make([][2]int64, len(pts))
	for i := range pts {
		deltas[i][0] = pts[i][0] - c.x
		deltas[i][1] = pts[i][1] - c.y
		c.x, c.y = pts[i][0], pts[i][1]
	}
	return deltas
}

func (c *cursor) encodeCmd(cmd uint32, points [][2]float64) []uint32 {
	if len(points) == 0 {
		return []uint32{}
	}
	// new slice to hold our encode bytes. 2 bytes for each point pluse a command byte.
	g := make([]uint32, 0, (2*len(points))+1)
	// add the command integer
	g = append(g, cmd)

	// range through our points
	for _, p := range points {
		dx, dy := c.GetDeltaPointAndUpdate(geom.Point(p))
		// encode our delta point
		g = append(g, encodeZigZag(dx), encodeZigZag(dy))
	}

	return g
}

// MoveTo encodes a move to command for the given points
func (c *cursor) MoveTo(points ...[2]float64) []uint32 {
	return c.encodeCmd(uint32(NewCommand(cmdMoveTo, len(points))), points)
}

// encodeGeometry will take a geom.Geometry and encode it according to the
// mapbox vector_tile spec.
func encodeGeometry(ctx context.Context, geometry geom.Geometry) (g []uint32, vtyp vectorTile.Tile_GeomType, err error) {
	if geometry == nil {
		return nil, vectorTile.Tile_UNKNOWN, ErrNilGeometryType
	}

	c := NewCursor()

	switch t := geometry.(type) {
	case geom.Point:
		g = append(g, c.MoveTo(t)...)
		return g, vectorTile.Tile_POINT, nil

	case geom.MultiPoint:
		g = append(g, c.MoveTo(t.Points()...)...)
		return g, vectorTile.Tile_POINT, nil

	default:
		return nil, vectorTile.Tile_UNKNOWN, ErrUnknownGeometryType
	}
}

// keyvalMapsFromFeatures returns a key map and value map, to help with the translation
// to mapbox tile format. In the Tile format, the Tile contains a mapping of all the unique
// keys and values, and then each feature contains a vector map to these two. This is an
// intermediate data structure to help with the construction of the three mappings.
func keyvalMapsFromFeatures(features []Feature) (keyMap []string, valMap []interface{}, err error) {
	var didFind bool
	for _, f := range features {
		for k, v := range f.Tags {
			didFind = false
			for _, mk := range keyMap {
				if k == mk {
					didFind = true
					break
				}
			}
			if !didFind {
				keyMap = append(keyMap, k)
			}
			didFind = false
			switch vt := v.(type) {
			default:
				if vt == nil {
					continue
				}
				return keyMap, valMap, fmt.Errorf("unsupported type for value(%v) with key(%v) in tags for feature %v.", vt, k, f)

			case string:
				for _, mv := range valMap {
					tmv, ok := mv.(string)
					if !ok {
						continue
					}
					if tmv == vt {
						didFind = true
						break
					}
				}
			case int:
				for _, mv := range valMap {
					tmv, ok := mv.(int)
					if !ok {
						continue
					}
					if tmv == vt {
						didFind = true
						break
					}
				}
			case float64:
				for _, mv := range valMap {
					tmv, ok := mv.(float64)
					if !ok {
						continue
					}
					if tmv == vt {
						didFind = true
						break
					}
				}
			}

			if !didFind {
				valMap = append(valMap, v)
			}

		}
	}
	return keyMap, valMap, nil
}

// keyvalTagsMap will return the tags map as expected by the mapbox tile spec. It takes
// a keyMap and a valueMap that list the the order of the expected keys and values. It will
// return a vector map that refers to these two maps.
func keyvalTagsMap(keyMap []string, valueMap []interface{}, f *Feature) (tags []uint32, err error) {
	if f == nil {
		return nil, ErrNilFeature
	}
	var kidx, vidx int64
	for key, val := range f.Tags {

		kidx, vidx = -1, -1

		for i, k := range keyMap {
			if k != key {
				continue
			}
			kidx = int64(i)
			break
		}

		if kidx == -1 {
			log.Printf("did not find key (%v) in keymap.", key)
			return tags, fmt.Errorf("did not find key (%v) in keymap.", key)
		}

		if val == nil {
			continue
		}

		for i, v := range valueMap {
			switch tv := val.(type) {
			default:
				return tags, fmt.Errorf("value (%[1]v) of type (%[1]T) for key (%[2]v) is not supported.", tv, key)
			case string:
				vmt, ok := v.(string)
				if !ok || vmt != tv {
					continue
				}
			case int:
				vmt, ok := v.(int)
				if !ok || vmt != tv {
					continue
				}
			case float64:
				vmt, ok := v.(float64)
				if !ok || vmt != tv {
					continue
				}
			case bool:
				vmt, ok := v.(bool)
				if !ok || vmt != tv {
					continue
				}
			}
			vidx = int64(i)
			break
		}

		if vidx == -1 { // None of the values matched.
			return tags, fmt.Errorf("did not find a value: %v in valuemap.", val)
		}
		tags = append(tags, uint32(kidx), uint32(vidx))
	}
	return tags, nil
}
