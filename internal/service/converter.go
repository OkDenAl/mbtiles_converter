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

// Converter represents converters
type Converter interface {
	Convert(ctx context.Context, opts config.ConverterOpts, meta config.Metadata) error
}

type converter struct {
	pgRepo     pg.Repository
	sqliteRepo sqliterepo.Repository
}

// NewConverter creates new converter
func NewConverter(pgRepo pg.Repository, sqliteRepo sqliterepo.Repository) Converter {
	return &converter{pgRepo: pgRepo, sqliteRepo: sqliteRepo}
}

// makeMapProjection makes projection from usual MapPoint with longitude and latitude to Tile
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

// makeTileDict creates a dictionary displaying tile coordinates to tiles points array
func makeTileDict(points []entity.MapPoint, startZoom, endZoom int) map[entity.TileCoords][]entity.TilePoint {
	tiles := make(map[entity.TileCoords][]entity.TilePoint, 0)
	for _, point := range points {
		for zoom := startZoom; zoom < endZoom; zoom++ {
			tile, pixelCoordInTile := makeMapProjection(point, zoom)
			if _, ok := tiles[tile]; !ok {
				tiles[tile] = make([]entity.TilePoint, 0)
			}
			tiles[tile] = append(tiles[tile], pixelCoordInTile)
		}
	}
	return tiles
}

// addNewPointsToMVT decodes outdated tile_data and adds info about a new tile point and encode it to mvt format again
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

// convertHelper converts the data
func (c *converter) convertHelper(ctx context.Context, points []entity.MapPoint, startZoom, endZoom int) error {
	tiles := makeTileDict(points, startZoom, endZoom)
	tileToAdd := make([]entity.MbtilesMapPoint, 0)
	tilesToUpdate := make([]entity.MbtilesMapPoint, 0)

	for tile, val := range tiles {
		mbtilesPoint := entity.MbtilesMapPoint{TileCol: tile.Column, TileRow: tile.Row, ZoomLevel: tile.Zoom}
		tileData, err := c.sqliteRepo.GetTileData(ctx, tile)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				mbtilesPoint.TileData, err = utils.EncodePixelCoordToGzipMVT(val, tile.Zoom)
				if err != nil {
					return fmt.Errorf("Convert-utils.EncodePixelCoordToGzipMVT : %w", err)
				}
				tileToAdd = append(tileToAdd, mbtilesPoint)
				continue
			}
			return fmt.Errorf("Convert-c.sqliteRepo.GetTileData : %w", err)
		}
		mbtilesPoint.TileData, err = addNewPointsToMVT(tileData, val)
		if err != nil {
			return fmt.Errorf("Convert-addNewPointsToMVT : %w", err)
		}
		tilesToUpdate = append(tilesToUpdate, mbtilesPoint)
	}

	if len(tileToAdd) > 0 {
		err := c.sqliteRepo.AddTilesBatch(ctx, tileToAdd)
		if err != nil {
			return fmt.Errorf("Convert-c.sqliteRepo.AddTilesBatch: %w", err)
		}
	}
	if len(tilesToUpdate) > 0 {
		err := c.sqliteRepo.UpdateTilesDataBatch(ctx, tilesToUpdate)
		if err != nil {
			return fmt.Errorf("Convert-c.sqliteRepo.UpdateTileData : %w", err)
		}
	}
	return nil
}

// Convert converts the geo object information from PostgreSQL to MBtiles vector format
func (c *converter) Convert(ctx context.Context, opts config.ConverterOpts, meta config.Metadata) error {
	err := c.sqliteRepo.CreateTables(ctx)
	if err != nil {
		return err
	}
	err = c.sqliteRepo.FillMetadata(ctx, entity.NewMetadata(meta))
	if err != nil {
		return err
	}
	offset := 0
	for offset < opts.ConvertLimit {
		points, err := c.pgRepo.GetNElements(ctx, opts.BatchSize, offset)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		err = c.convertHelper(ctx, points, opts.StartZoom, opts.EndZoom)
		if err != nil {
			return err
		}
		offset += opts.BatchSize
	}
	return nil
}
