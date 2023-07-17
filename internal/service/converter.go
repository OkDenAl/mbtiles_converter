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
			tile, coords := makeMapProjection(point, zoom)
			if _, ok := tiles[tile]; !ok {
				tiles[tile] = make([][2]float64, 0)
			}
			tiles[tile] = append(tiles[tile], coords)
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

func (c *converter) convertWithMapAndSqlite(ctx context.Context, points []entity.MapPoint, startZoom, endZoom int) error {
	tiles := make(map[entity.TileCoords][][2]float64, 0)
	mbtilesPoints := make([]entity.MbtilesMapPoint, 0)
	for _, point := range points {
		for zoom := startZoom; zoom < endZoom; zoom++ {
			tile, coords := makeMapProjection(point, zoom)
			if _, ok := tiles[tile]; !ok {
				tiles[tile] = make([][2]float64, 0)
			}
			tiles[tile] = append(tiles[tile], coords)
		}
	}

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

func makeMapProjection(point entity.MapPoint, zoom int) (entity.TileCoords, [2]float64) {
	tile := entity.TileCoords{
		Column: utils.Lon2tileFloor(point.Longitude, zoom),
		Row:    utils.Lat2tileFloor(point.Latitude, zoom),
		Zoom:   zoom,
	}
	tileSize := 1 << zoom
	x := math.Round((tile.Column - utils.Lon2tile(point.Longitude, zoom)) * float64(-tileSize))
	y := math.Round((tile.Row - utils.Lat2tile(point.Latitude, zoom)) * float64(-tileSize))
	return tile, [2]float64{x, y}
}

func addNewPointsToMVT(tileData []byte, val [][2]float64) ([]byte, error) {
	decodedTile, err := utils.DecodeFromGzipMVT(tileData)
	if err != nil {
		return nil, fmt.Errorf("utils.DecodeFromGzipMVT : %w", err)
	}
	geo := geom.MultiPoint{}
	_ = geo.SetPoints(val)
	f := mvt.NewFeatures(geo, nil)
	decodedTile.Layers[0] = decodedTile.TakeLayers()[0].AddFeatures(f...)
	tileData, err = utils.EncodeTileToMVT(*decodedTile)
	if err != nil {
		return nil, fmt.Errorf("utils.EncodeTileToMVT : %w", err)
	}
	return tileData, nil
}

func (c *converter) Convert(ctx context.Context, opts config.ConverterOpts) error {
	offset := 0
	for offset < 10000 {
		points, err := c.pgRepo.GetNElements(ctx, opts.QuantityToConvert, offset)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		switch opts.ConverterMode {
		case 1:
			err = c.convertUsingMap(ctx, points, opts.StartZoom, opts.EndZoom)
		case 2:
			err = c.convertWithMapAndSqlite(ctx, points, opts.StartZoom, opts.EndZoom)
		default:
			err = ErrInvalidConvertMode
		}
		offset += opts.QuantityToConvert
		if err != nil {
			return err
		}
	}
	return nil
}
