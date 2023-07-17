package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/config"
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/pg"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/sqliterepo"
	"github.com/OkDenAl/mbtiles_converter/pkg/mvt"
	"github.com/OkDenAl/mbtiles_converter/pkg/utils"
	"github.com/go-spatial/geom"
	"math"
)

var (
	ErrInvalidConvertMode = errors.New("invalid convert mode")
)

type Converter interface {
	Convert(ctx context.Context, opts config.ConverterOpts) error
}

type converter struct {
	pgRepo     pg.Repository
	sqliteRepo sqliterepo.Repository
}

func NewConverter(pgRepo pg.Repository, sqliteRepo sqliterepo.Repository) Converter {
	return &converter{pgRepo: pgRepo, sqliteRepo: sqliteRepo}
}

func makeMapProjection(point entity.MapPoint, zoom int) (entity.TileCoords, entity.TilePoint) {
	tile := entity.TileCoords{
		Column: utils.Lon2tileFloor(point.Longitude, zoom),
		Row:    utils.Lat2tileFloor(point.Latitude, zoom),
		Zoom:   zoom,
	}

	tileSize := 1 << zoom
	tilePoint := entity.TilePoint{
		X:    math.Round((tile.Column - utils.Lon2tile(point.Longitude, zoom)) * float64(-tileSize)),
		Y:    math.Round((tile.Row - utils.Lat2tile(point.Latitude, zoom)) * float64(-tileSize)),
		Type: point.Type,
	}
	return tile, tilePoint
}

func makeTileDict(points []entity.MapPoint, startZoom, endZoom int) map[entity.TileCoords][]entity.TilePoint {
	tiles := make(map[entity.TileCoords][]entity.TilePoint, 0)
	for _, point := range points {
		for zoom := startZoom; zoom < endZoom; zoom++ {
			tile, coords := makeMapProjection(point, zoom)
			if _, ok := tiles[tile]; !ok {
				tiles[tile] = make([]entity.TilePoint, 0)
			}
			tiles[tile] = append(tiles[tile], coords)
		}
	}
	return tiles
}

func addNewPointsToMVT(tileData []byte, tilePoints []entity.TilePoint) ([]byte, error) {
	decodedTile, err := utils.DecodeFromGzipMVT(tileData)
	if err != nil {
		return nil, fmt.Errorf("utils.DecodeFromGzipMVT : %w", err)
	}
	features := make([]mvt.Feature, len(tilePoints))
	for i, tilePoint := range tilePoints {
		geo := geom.Point{tilePoint.X, tilePoint.Y}
		features[i] = mvt.Feature{Geometry: geo, Tags: map[string]interface{}{"type": tilePoint.Type}}
	}
	decodedTile.Layers[0] = decodedTile.TakeLayers()[0].AddFeatures(features...)
	tileData, err = utils.EncodeTileToMVT(*decodedTile)
	if err != nil {
		return nil, fmt.Errorf("utils.EncodeTileToMVT : %w", err)
	}
	return tileData, nil
}

func (c *converter) convert(ctx context.Context, points []entity.MapPoint, startZoom, endZoom int) error {
	tiles := makeTileDict(points, startZoom, endZoom)
	mbtilesPoints := make([]entity.MbtilesMapPoint, 0)

	for tile, val := range tiles {
		mbtilesPoint := entity.MbtilesMapPoint{TileCol: tile.Column, TileRow: tile.Row, ZoomLevel: tile.Zoom}
		tileData, err := c.sqliteRepo.GetTileData(ctx, tile)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				mbtilesPoint.TileData, err = utils.EncodePixelCoordToGzipMVT(val, tile.Zoom)
				if err != nil {
					return fmt.Errorf("Convert-utils.EncodePixelCoordToGzipMVT : %w", err)
				}
				mbtilesPoints = append(mbtilesPoints, mbtilesPoint)
				continue
			}
			return fmt.Errorf("Convert-c.sqliteRepo.GetTileData : %w", err)
		}
		mbtilesPoint.TileData, err = addNewPointsToMVT(tileData, val)
		if err != nil {
			return fmt.Errorf("Convert-addNewPointsToMVT : %w", err)
		}
		err = c.sqliteRepo.UpdateTileData(ctx, mbtilesPoint)
		if err != nil {
			return fmt.Errorf("Convert-c.sqliteRepo.UpdateTileData : %w", err)
		}
	}

	if len(mbtilesPoints) > 0 {
		err := c.sqliteRepo.AddTilesBatch(ctx, mbtilesPoints)
		if err != nil {
			return fmt.Errorf("Convert-c.sqliteRepo.AddTilesBatch: %w", err)
		}
	}
	return nil
}

func (c *converter) Convert(ctx context.Context, opts config.ConverterOpts) error {
	offset := 0
	for {
		points, err := c.pgRepo.GetNElements(ctx, opts.QuantityToConvert, offset)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		err = c.convert(ctx, points, opts.StartZoom, opts.EndZoom)
		if err != nil {
			return err
		}
		offset += opts.QuantityToConvert
	}
	//return nil
}
