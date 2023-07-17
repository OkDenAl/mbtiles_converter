package sqliterepo

import (
	"context"
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"github.com/OkDenAl/mbtiles_converter/pkg/sqlite"
	"log"
	"strings"
)

type Repository interface {
	AddTilesBatch(ctx context.Context, mbtilesPoints []entity.MbtilesMapPoint) error
	CreateTables(ctx context.Context) error
	GetTileData(ctx context.Context, tile entity.TileCoords) ([]byte, error)
	AddTile(ctx context.Context, point entity.MbtilesMapPoint) error
	UpdateTileData(ctx context.Context, point entity.MbtilesMapPoint) error
	FillMetadata(ctx context.Context, metadata entity.Metadata) error
}

type repo struct {
	conn *sqlite.Pool
}

func NewRepo(conn *sqlite.Pool) Repository {
	return &repo{conn: conn}
}

func (r *repo) AddTilesBatch(ctx context.Context, mbtilesPoints []entity.MbtilesMapPoint) error {
	valueStrings := make([]string, 0, len(mbtilesPoints))
	valueArgs := make([]interface{}, 0, len(mbtilesPoints)*4)
	for i, point := range mbtilesPoints {
		maxZ := 1 << point.ZoomLevel
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d,$%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		valueArgs = append(valueArgs, point.ZoomLevel)
		valueArgs = append(valueArgs, point.TileCol)
		valueArgs = append(valueArgs, float64(maxZ)-point.TileRow-1)
		valueArgs = append(valueArgs, point.TileData)
	}
	stmt := fmt.Sprintf("INSERT INTO tiles (zoom_level, tile_column, tile_row , tile_data) VALUES %s",
		strings.Join(valueStrings, ","))
	log.Println(stmt)
	db := r.conn.Checkout()
	defer r.conn.Checkin(db)
	_, err := db.ExecContext(ctx, stmt, valueArgs...)
	return err
}

func (r *repo) CreateTables(ctx context.Context) error {
	db := r.conn.Checkout()
	defer r.conn.Checkin(db)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	q := `CREATE TABLE IF NOT EXISTS metadata (name text, value text);`
	_, err = tx.Exec(q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS tiles (zoom_level integer, tile_column integer, tile_row integer, tile_data blob);`
	_, err = tx.Exec(q)
	if err != nil {
		return err
	}
	q = `CREATE UNIQUE INDEX IF NOT EXISTS tile_index on tiles (zoom_level, tile_column, tile_row)`
	_, err = tx.Exec(q)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *repo) GetTileData(ctx context.Context, tile entity.TileCoords) ([]byte, error) {
	db := r.conn.Checkout()
	defer r.conn.Checkin(db)
	maxZ := 1 << tile.Zoom
	q := `SELECT tile_data FROM tiles WHERE zoom_level = $1 AND tile_column = $2 AND tile_row = $3`
	row := db.QueryRowContext(ctx, q, tile.Zoom, tile.Column, float64(maxZ)-tile.Row-1)
	var tileData []byte
	err := row.Scan(&tileData)
	if err != nil {
		return nil, err
	}
	return tileData, nil
}

func (r *repo) AddTile(ctx context.Context, point entity.MbtilesMapPoint) error {
	db := r.conn.Checkout()
	defer r.conn.Checkin(db)
	maxZ := 1 << point.ZoomLevel
	q := `INSERT INTO tiles (zoom_level, tile_column, tile_row , tile_data) VALUES ($1,$2,$3,$4)`
	_, err := db.ExecContext(ctx, q, point.ZoomLevel, point.TileCol, float64(maxZ)-point.TileRow-1, point.TileData)
	return err
}

func (r *repo) UpdateTileData(ctx context.Context, point entity.MbtilesMapPoint) error {
	db := r.conn.Checkout()
	defer r.conn.Checkin(db)
	maxZ := 1 << point.ZoomLevel
	q := `UPDATE tiles SET tile_data = $1 WHERE zoom_level = $2 AND tile_column = $3 AND tile_row = $4`
	_, err := db.ExecContext(ctx, q, point.TileData, point.ZoomLevel, point.TileCol, float64(maxZ)-point.TileRow-1)
	return err
}

func (r *repo) FillMetadata(ctx context.Context, metadata entity.Metadata) error {
	log.Println("starting to fill metadata")
	return nil
}
