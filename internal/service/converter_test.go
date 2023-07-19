package service

import (
	"context"
	"database/sql"
	"github.com/OkDenAl/mbtiles_converter/config"
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"github.com/OkDenAl/mbtiles_converter/internal/service/pg_mocks"
	"github.com/OkDenAl/mbtiles_converter/internal/service/sqlite_mocks"
	"github.com/OkDenAl/mbtiles_converter/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestMakeMapProjection(t *testing.T) {
	coords, point := makeMapProjection(entity.MapPoint{Latitude: 55.0, Longitude: 30.5, Type: "cafe"}, 6)
	assert.Equal(t, coords, entity.TileCoords{Column: 37, Row: 20, Zoom: 6})
	assert.Equal(t, point, entity.TilePoint{X: 27, Y: 16, Type: "cafe"})
}

func TestMakeTileDict(t *testing.T) {
	hashTable := makeTileDict([]entity.MapPoint{{Latitude: 55.0, Longitude: 30.5, Type: "cafe"}}, 6, 8)
	assert.Equal(t, hashTable[entity.TileCoords{Column: 37, Row: 20, Zoom: 6}][0], entity.TilePoint{X: 27, Y: 16, Type: "cafe"})
	assert.Len(t, hashTable, 2)
}

func TestAddNewPointsToMVT(t *testing.T) {
	mvt, err := utils.EncodePixelCoordToGzipMVT([]entity.TilePoint{{X: 10, Y: 9, Type: "cafe"}}, 6)
	assert.NoError(t, err)
	toMVT, err := addNewPointsToMVT(mvt, []entity.TilePoint{{X: 27, Y: 16, Type: "cafe"}})
	assert.NoError(t, err)
	assert.Len(t, toMVT, 76)
	a, err := utils.DecodeFromGzipMVT(toMVT)
	assert.NoError(t, err)
	assert.Len(t, a.TakeLayers(), 1)
	assert.Len(t, a.TakeLayers()[0].Features, 2)
}

func TestConvert_OkAddTilesBatch(t *testing.T) {
	sqliteRepo := &sqlite_mocks.Repository{}

	sqliteRepo.On("CreateTables", mock.AnythingOfType("*context.emptyCtx")).Return(nil)
	sqliteRepo.On("FillMetadata", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("entity.Metadata")).Return(nil)
	sqliteRepo.On("GetTileData", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("entity.TileCoords")).Return(nil, sql.ErrNoRows)
	sqliteRepo.On("AddTilesBatch", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("[]entity.MbtilesMapPoint")).Return(nil)

	pgRepo := &pg_mocks.Repository{}
	pgRepo.On("GetNElements", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return([]entity.MapPoint{{Latitude: 55.0, Longitude: 30.5, Type: "cafe"}}, nil)

	conv := NewConverter(pgRepo, sqliteRepo)
	err := conv.Convert(context.Background(),
		config.ConverterOpts{BatchSize: 10, ConvertLimit: 10, StartZoom: 5, EndZoom: 7},
		config.Metadata{},
	)
	assert.NoError(t, err)
}

func TestConvert_NoRowsInPG(t *testing.T) {
	sqliteRepo := &sqlite_mocks.Repository{}
	sqliteRepo.On("CreateTables", mock.AnythingOfType("*context.emptyCtx")).Return(nil)
	sqliteRepo.On("FillMetadata", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("entity.Metadata")).Return(nil)

	pgRepo := &pg_mocks.Repository{}
	pgRepo.On("GetNElements", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(nil, sql.ErrNoRows)

	conv := NewConverter(pgRepo, sqliteRepo)
	err := conv.Convert(context.Background(),
		config.ConverterOpts{BatchSize: 10, ConvertLimit: 10, StartZoom: 5, EndZoom: 7},
		config.Metadata{},
	)
	assert.NoError(t, err)
}

func TestConvert_OkUpdateTilesBatch(t *testing.T) {
	sqliteRepo := &sqlite_mocks.Repository{}

	sqliteRepo.On("CreateTables", mock.AnythingOfType("*context.emptyCtx")).Return(nil)
	sqliteRepo.On("FillMetadata", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("entity.Metadata")).Return(nil)
	mvt, err := utils.EncodePixelCoordToGzipMVT([]entity.TilePoint{{X: 10, Y: 9, Type: "cafe"}}, 6)
	assert.NoError(t, err)
	sqliteRepo.On("GetTileData", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("entity.TileCoords")).Return(mvt, nil)
	sqliteRepo.On("UpdateTilesDataBatch", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("[]entity.MbtilesMapPoint")).Return(nil)

	pgRepo := &pg_mocks.Repository{}
	pgRepo.On("GetNElements", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return([]entity.MapPoint{{Latitude: 55.0, Longitude: 30.5, Type: "cafe"}}, nil)

	conv := NewConverter(pgRepo, sqliteRepo)
	err = conv.Convert(context.Background(),
		config.ConverterOpts{BatchSize: 10, ConvertLimit: 10, StartZoom: 5, EndZoom: 7},
		config.Metadata{},
	)
	assert.NoError(t, err)
}
