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

func (c *converter) convertUsingMap(ctx context.Context, points []entity.MapPoint, startZoom, endZoom int) error {
	tiles := make(map[entity.TileCoords][][2]float64, 0)
	mbtilesPoints := make([]entity.MbtilesMapPoint, 0)
	for _, point := range points {
		for zoom := startZoom; zoom < endZoom; zoom++ {
			tile := entity.TileCoords{
				Column: utils.Lon2tileFloor(point.Longitude, zoom),
				Row:    utils.Lat2tileFloor(point.Latitude, zoom),
				Zoom:   zoom,
			}
			tileSize := 1 << zoom
			x := math.Round((tile.Column - utils.Lon2tile(point.Longitude, zoom)) * float64(-tileSize))
			y := math.Round((tile.Row - utils.Lat2tile(point.Latitude, zoom)) * float64(-tileSize))
			if _, ok := tiles[tile]; !ok {
				tiles[tile] = make([][2]float64, 0)
			}
			tiles[tile] = append(tiles[tile], [2]float64{x, y})
		}
	}
	for tile, val := range tiles {
		gzipBuf, err := utils.EncodePixelCoordToGzipMVT(val, tile.Zoom)
		if err != nil {
			return fmt.Errorf("Convert-utils.EncodePixelCoordToGzipMVT: %w", err)
		}
		mbtilesPoints = append(mbtilesPoints, entity.MbtilesMapPoint{TileCol: tile.Column,
			TileRow: tile.Row, ZoomLevel: tile.Zoom, TileData: gzipBuf})
	}
	err := c.sqliteRepo.AddTilesBatch(ctx, mbtilesPoints)
	if err != nil {
		return fmt.Errorf("Convert-c.sqliteRepo.AddTilesBatch: %w", err)
	}
	return nil
}

func (c *converter) convertWithoutMap(ctx context.Context, points []entity.MapPoint, startZoom, endZoom int) error {
	for _, point := range points {
		for zoom := startZoom; zoom < endZoom; zoom++ {
			tile := entity.TileCoords{
				Column: utils.Lon2tileFloor(point.Longitude, zoom),
				Row:    utils.Lat2tileFloor(point.Latitude, zoom),
				Zoom:   zoom,
			}
			tileSize := 1 << zoom
			x := math.Round((tile.Column - utils.Lon2tile(point.Longitude, zoom)) * float64(-tileSize))
			y := math.Round((tile.Row - utils.Lat2tile(point.Latitude, zoom)) * float64(-tileSize))

			mbtilesPoint := entity.MbtilesMapPoint{TileCol: tile.Column, TileRow: tile.Row, ZoomLevel: tile.Zoom}
			tileData, err := c.sqliteRepo.GetTileData(ctx, tile)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					mbtilesPoint.TileData, err = utils.EncodePixelCoordToGzipMVT([][2]float64{{x, y}}, tile.Zoom)
					if err != nil {
						return fmt.Errorf("Convert-utils.EncodePixelCoordToGzipMVT : %w", err)
					}
					err = c.sqliteRepo.AddTile(ctx, mbtilesPoint)
					if err != nil {
						return fmt.Errorf("Convert-c.sqliteRepo.AddTile : %w", err)
					}
					continue
				}
				return fmt.Errorf("Convert-c.sqliteRepo.GetTileData : %w", err)
			}

			decodedTile, err := utils.DecodeFromGzipMVT(tileData)
			if err != nil {
				return fmt.Errorf("Convert-utils.DecodeFromGzipMVT : %w", err)
			}
			f := mvt.NewFeatures(geom.Point{x, y}, nil)
			decodedTile.Layers[0] = decodedTile.TakeLayers()[0].AddFeatures(f...)
			mbtilesPoint.TileData, err = utils.EncodeTileToMVT(*decodedTile)
			if err != nil {
				return fmt.Errorf("Convert-utils.EncodeTileToMVT : %w", err)
			}
			err = c.sqliteRepo.UpdateTileData(ctx, mbtilesPoint)
			if err != nil {
				return fmt.Errorf("Convert-c.sqliteRepo.UpdateTileData : %w", err)
			}
		}
	}
	return nil
}

func (c *converter) Convert(ctx context.Context, opts config.ConverterOpts) error {
	points, err := c.pgRepo.GetFirstNElements(ctx, opts.QuantityToConvert)
	if err != nil {
		return err
	}
	switch opts.ConverterMode {
	case 1:
		err = c.convertUsingMap(ctx, points, opts.StartZoom, opts.EndZoom)
	case 2:
		err = c.convertWithoutMap(ctx, points, opts.StartZoom, opts.EndZoom)
	default:
		err = ErrInvalidConvertMode
	}
	return err
}
