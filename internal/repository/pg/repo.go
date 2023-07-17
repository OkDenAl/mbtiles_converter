package pg

import (
	"context"
	"github.com/OkDenAl/mbtiles_converter/internal/entity"
	"github.com/OkDenAl/mbtiles_converter/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetNElements(ctx context.Context, n, offset int) ([]entity.MapPoint, error)
}

type repo struct {
	conn postgres.PgxPool
}

func NewRepo(conn *pgxpool.Pool) Repository {
	return &repo{conn: conn}
}

func (r *repo) GetNElements(ctx context.Context, n, offset int) ([]entity.MapPoint, error) {
	q := `SELECT longitude,latitude,type FROM geo_objects LIMIT $1 OFFSET $2`
	rows, err := r.conn.Query(ctx, q, n, offset)
	if err != nil {
		return nil, err
	}
	points := make([]entity.MapPoint, n)
	c := 0
	for rows.Next() {
		var point entity.MapPoint
		err = rows.Scan(&point.Longitude, &point.Latitude, &point.Type)
		if err != nil {
			return nil, err
		}
		points[c] = point
		c++
	}
	return points, nil
}
