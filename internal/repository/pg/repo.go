package pg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"github.com/OkDenAl/mbtiles_converter/pkg/postgres"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

// Repository represent the methods for PostgreSQL database
type Repository interface {
	// GetNElements returns N elements from PostgreSQL table with offset
	GetNElements(ctx context.Context, tableName, rowsNames string, n, offset int) ([]entity.MapPoint, error)
}

type repo struct {
	conn postgres.PgxPool
}

// NewRepo creates a new PostgreSQL repository
func NewRepo(conn *pgxpool.Pool) Repository {
	return &repo{conn: conn}
}

func (r *repo) GetNElements(ctx context.Context, tableName, rowsNames string, n, offset int) ([]entity.MapPoint, error) {
	q := fmt.Sprintf("SELECT %s FROM %s LIMIT $1 OFFSET $2", rowsNames, tableName)
	rows, err := r.conn.Query(ctx, q, n, offset)
	if err != nil {
		return nil, fmt.Errorf("r.conn.Query with query %s: %w", q, err)
	}
	points := make([]entity.MapPoint, n)
	c := 0
	rowsNamesArr := strings.Split(rowsNames, ",")[2:]
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("rows.Values: %w", err)
		}
		lon, _ := values[0].(pgtype.Numeric).Float64Value()
		lat, _ := values[1].(pgtype.Numeric).Float64Value()
		point := entity.MapPoint{Longitude: lon.Float64, Latitude: lat.Float64, AdditionalRows: make(map[string]any)}
		values = values[2:]
		for i, rowName := range rowsNamesArr {
			point.AdditionalRows[rowName] = values[i]
		}
		points[c] = point
		c++
	}
	if c == 0 {
		return nil, sql.ErrNoRows
	}
	return points, nil
}
