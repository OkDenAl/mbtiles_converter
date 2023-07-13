package service

import (
	"context"
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/pg"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/sqlite"
	"github.com/OkDenAl/mbtiles_converter/pkg/utils"
	"math"
)

type Converter interface {
	Convert(ctx context.Context, n int) error
}

type converter struct {
	pgRepo     pg.Repository
	sqliteRepo sqlite.Repository
}

func NewConverter(pgRepo pg.Repository, sqliteRepo sqlite.Repository) Converter {
	return &converter{pgRepo: pgRepo, sqliteRepo: sqliteRepo}
}

func (c *converter) Convert(ctx context.Context, n int) error {
	points, err := c.pgRepo.GetFirstNElements(ctx, n)
	if err != nil {
		return err
	}
	startZoom := 7
	endZoom := 12
	tiles := make(map[[3]int][][2]float64, 0)
	mbtilesPoints := make([]entity.MbtilesMapPoint, 0)
	for _, point := range points {
		for i := startZoom; i < endZoom; i++ {
			tileColumn := utils.Lon2tileFloor(point.Longitude, i)
			tileRow := utils.Lat2tileFloor(point.Latitude, i)
			tileSize := 1 << i
			x := math.Round((tileColumn - utils.Lon2tile(point.Longitude, i)) * float64(-tileSize))
			y := math.Round((tileRow - utils.Lat2tile(point.Latitude, i)) * float64(-tileSize))
			if _, ok := tiles[[3]int{int(tileColumn), int(tileRow), i}]; !ok {
				tiles[[3]int{int(tileColumn), int(tileRow), i}] = make([][2]float64, 0)
			}
			tiles[[3]int{int(tileColumn), int(tileRow), i}] = append(tiles[[3]int{int(tileColumn), int(tileRow), i}], [2]float64{x, y})
		}
	}
	for k, val := range tiles {
		gzipBuf, err := utils.ConvertToMVT(val, k[2])
		if err != nil {
			return err
		}
		mbtilesPoints = append(mbtilesPoints, entity.MbtilesMapPoint{TileCol: float64(k[0]),
			TileRow: float64(k[1]), ZoomLevel: k[2], TileData: gzipBuf})
	}
	err = c.sqliteRepo.AddPoint(ctx, mbtilesPoints)
	if err != nil {
		return err
	}
	return nil
}
