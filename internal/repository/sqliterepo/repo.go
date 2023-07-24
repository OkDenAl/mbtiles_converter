package sqliterepo

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"reflect"
	"strings"
)

// Repository represent the methods for SQLite database
type Repository interface {
	// AddTilesBatch adds data about tiles in one sql query
	AddTilesBatch(ctx context.Context, mbtilesPoints []entity.MbtilesMapPoint) error
	// CreateTables creates a tables: metadata and tiles (read more in MBtiles spec)
	CreateTables(ctx context.Context) error
	// GetTileData returns the tile_data column from tiles table
	GetTileData(ctx context.Context, tile entity.TileCoords) ([]byte, error)
	// AddTile adds data about one tile
	AddTile(ctx context.Context, point entity.MbtilesMapPoint) error
	// UpdateTileData updates one tile_data column
	UpdateTileData(ctx context.Context, point entity.MbtilesMapPoint) error
	// UpdateTilesDataBatch updates tile_data columns in one sql transaction
	UpdateTilesDataBatch(ctx context.Context, mbtilesPoints []entity.MbtilesMapPoint) error
	// FillMetadata fills the metadata table
	FillMetadata(ctx context.Context, metadata entity.Metadata) error
}

type repo struct {
	conn *sql.DB
}

func NewRepo(conn *sql.DB) Repository {
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
	_, err := r.conn.ExecContext(ctx, stmt, valueArgs...)
	if err != nil {
		return fmt.Errorf("r.conn.ExecContext with query `%s`: %w", stmt, err)
	}
	return fmt.Errorf("r.conn.ExecContext: %w", err)
}

func (r *repo) CreateTables(ctx context.Context) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("r.conn.BeginTx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	q := `CREATE TABLE IF NOT EXISTS metadata (name text, value text);`
	_, err = tx.Exec(q)
	if err != nil {
		return fmt.Errorf("tx.Exec with query `%s`: %w", q, err)
	}
	q = `CREATE TABLE IF NOT EXISTS tiles (zoom_level integer, tile_column integer, tile_row integer, tile_data blob);`
	_, err = tx.Exec(q)
	if err != nil {
		return fmt.Errorf("tx.Exec with query `%s`: %w", q, err)
	}
	q = `CREATE UNIQUE INDEX IF NOT EXISTS tile_index on tiles (zoom_level, tile_column, tile_row)`
	_, err = tx.Exec(q)
	if err != nil {
		return fmt.Errorf("tx.Exec with query `%s`: %w", q, err)
	}
	return tx.Commit()
}

func (r *repo) GetTileData(ctx context.Context, tile entity.TileCoords) ([]byte, error) {
	maxZ := 1 << tile.Zoom
	q := `SELECT tile_data FROM tiles WHERE zoom_level = $1 AND tile_column = $2 AND tile_row = $3`
	row := r.conn.QueryRowContext(ctx, q, tile.Zoom, tile.Column, float64(maxZ)-tile.Row-1)
	var tileData []byte
	err := row.Scan(&tileData)
	return tileData, err
}

func (r *repo) AddTile(ctx context.Context, point entity.MbtilesMapPoint) error {
	maxZ := 1 << point.ZoomLevel
	q := `INSERT INTO tiles (zoom_level, tile_column, tile_row , tile_data) VALUES ($1,$2,$3,$4)`
	_, err := r.conn.ExecContext(ctx, q, point.ZoomLevel, point.TileCol, float64(maxZ)-point.TileRow-1, point.TileData)
	if err != nil {
		return fmt.Errorf("r.conn.ExecContext with query `%s`: %w", q, err)
	}
	return nil
}

func (r *repo) UpdateTileData(ctx context.Context, point entity.MbtilesMapPoint) error {
	maxZ := 1 << point.ZoomLevel
	q := `UPDATE tiles SET tile_data = $1 WHERE zoom_level = $2 AND tile_column = $3 AND tile_row = $4`
	_, err := r.conn.ExecContext(ctx, q, point.TileData, point.ZoomLevel, point.TileCol, float64(maxZ)-point.TileRow-1)
	if err != nil {
		return fmt.Errorf("r.conn.ExecContext with query `%s`: %w", q, err)
	}
	return nil
}

func (r *repo) UpdateTilesDataBatch(ctx context.Context, mbtilesPoints []entity.MbtilesMapPoint) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("r.conn.BeginTx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	q := `UPDATE tiles SET tile_data = $1 WHERE zoom_level = $2 AND tile_column = $3 AND tile_row = $4`
	for _, point := range mbtilesPoints {
		_, err = tx.ExecContext(ctx, q, point.TileData, point.ZoomLevel, point.TileCol, point.TileRow)
		if err != nil {
			return fmt.Errorf("r.conn.ExecContext with query `%s`: %w", q, err)
		}
	}
	return tx.Commit()
}

func (r *repo) FillMetadata(ctx context.Context, metadata entity.Metadata) error {
	v := reflect.ValueOf(metadata)
	typeOfM := v.Type()
	valueStrings := make([]string, 0, v.NumField())
	valueArgs := make([]interface{}, 0, v.NumField()*2)
	for i := 0; i < v.NumField(); i++ {
		switch v.Field(i).Interface().(type) {
		case int:
			if v.Field(i).Interface() == 0 && typeOfM.Field(i).Name != "MinZoom" {
				continue
			}
		case string:
			if v.Field(i).Interface() == "" {
				continue
			}
		}
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, strings.ToLower(typeOfM.Field(i).Name))
		valueArgs = append(valueArgs, v.Field(i).Interface())
	}
	stmt := fmt.Sprintf("INSERT INTO metadata (name, value) VALUES %s", strings.Join(valueStrings, ","))
	_, err := r.conn.ExecContext(ctx, stmt, valueArgs...)
	if err != nil {
		return fmt.Errorf("r.conn.ExecContext with query `%s`: %w", stmt, err)
	}
	return nil
}
