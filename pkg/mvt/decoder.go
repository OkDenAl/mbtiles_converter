package mvt

import (
	"errors"
	"fmt"
	vectorTile "github.com/OkDenAl/mbtiles_converter/pkg/mvt/vector_tile"
	"github.com/arolek/p"
	"github.com/go-spatial/geom"
	"google.golang.org/protobuf/proto"
)

var (
	ErrExtraData   = errors.New("mvt: invalid extra data")
	ErrUnknownType = errors.New("mvt: unknown type")
)

// DecodeByte decodes the MVT encoded bytes into a Tile.
func DecodeByte(b []byte) (*Tile, error) {
	vtile := new(vectorTile.Tile)

	err := proto.Unmarshal(b, vtile)
	if err != nil {
		return nil, err
	}

	ret := new(Tile)
	ret.Layers = make([]Layer, len(vtile.Layers))

	for i, v := range vtile.Layers {
		err = decodeLayer(v, &ret.Layers[i])
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func decodeLayer(pb *vectorTile.Tile_Layer, dst *Layer) error {
	dst.Name = pb.Name
	dst.extent = p.Int(int(pb.Extent))

	dst.Features = make([]Feature, len(pb.Features))

	for i, v := range pb.Features {
		err := decodeFeature(v, &dst.Features[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func decodeFeature(pb *vectorTile.Tile_Feature, dst *Feature) error {
	dst.ID = &pb.Id
	var err error
	dst.Geometry, err = DecodeGeometry(pb.Type, pb.Geometry)
	return err
}

func DecodeGeometry(gtype vectorTile.Tile_GeomType, b []uint32) (geom.Geometry, error) {
	switch gtype {
	case vectorTile.Tile_POINT:
		return decodePoint(b)
	default:
		return nil, ErrUnknownGeometryType
	}
}

func decodePoint(buf []uint32) (geom.Geometry, error) {
	ret := [][2]float64{}
	curs := decodeCursor{}

	if len(buf) > 0 {
		cmd := Command(buf[0])
		buf = buf[1:]

		if len(buf) < cmd.Count()*2 {
			return nil, fmt.Errorf("not enough integers (%v) for %d", len(buf), cmd)
		}

		switch cmd.ID() {
		case cmdMoveTo:
			ret = curs.decodeNPoints(cmd.Count(), buf, false)
			buf = buf[cmd.Count()*2:]

		default:
			return nil, fmt.Errorf("invalid command for POINT, %d", cmd)
		}
	}

	if len(buf) != 0 {
		fmt.Println(buf)
		return ret, ErrExtraData
	}

	switch len(ret) {
	case 0:
		return nil, nil
	case 1:
		return geom.Point(ret[0]), nil
	default:
		return geom.MultiPoint(ret), nil
	}
}

type decodeCursor struct {
	x, y float64
}

func (c *decodeCursor) decodeNPoints(n int, pts []uint32, encHere bool) [][2]float64 {
	nd := 0

	if encHere {
		nd = 1
	}

	ret := make([][2]float64, n+nd)

	if encHere {
		ret[0] = [2]float64{c.x, c.y}
	}

	for i := 0; i < n; i++ {
		ret[i+nd] = c.decodePoint(pts[i*2], pts[i*2+1])
	}

	return ret
}

func (c *decodeCursor) decodePoint(x, y uint32) [2]float64 {
	c.x += float64(decodeZigZag(x))
	c.y += float64(decodeZigZag(y))
	return [2]float64{c.x, c.y}
}

func decodeZigZag(i uint32) int32 {
	return int32((i >> 1) ^ (-(i & 1)))
}
