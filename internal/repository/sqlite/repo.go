package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"strings"
)

type Repository interface {
	AddPoint(ctx context.Context, mbtilesPoints []entity.MbtilesMapPoint) error
	CreateTables(ctx context.Context) error
	FillMetadata(ctx context.Context, metadata entity.Metadata) error
}

type repo struct {
	conn *sql.DB
}

func NewRepo(conn *sql.DB) Repository {
	return &repo{conn: conn}
}

func (r *repo) AddPoint(ctx context.Context, mbtilesPoints []entity.MbtilesMapPoint) error {
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
	_, err := r.conn.Exec(stmt, valueArgs...)
	return err
}

func (r *repo) CreateTables(ctx context.Context) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	q := `CREATE TABLE metadata (name text, value text);`
	_, err = tx.Exec(q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE tiles (zoom_level integer, tile_column integer, tile_row integer, tile_data blob);`
	_, err = tx.Exec(q)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *repo) FillMetadata(ctx context.Context, metadata entity.Metadata) error {
	return nil
}
