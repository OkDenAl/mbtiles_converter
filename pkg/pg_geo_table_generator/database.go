package pg_geo_table_generator

import (
	"context"
	"github.com/OkDenAl/mbtiles_converter/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateTable(ctx context.Context) error
	FillTable(ctx context.Context, bord Borders, amount int) error
}

type repo struct {
	conn postgres.PgxPool
}

func NewRepo(conn *pgxpool.Pool) Repository {
	return &repo{conn: conn}
}

func (r *repo) CreateTable(ctx context.Context) error {
	q := `CREATE TABLE if NOT EXISTS geo_objects (
    	id SERIAL PRIMARY KEY,
    	longitude NUMERIC(10,4),
    	latitude NUMERIC (18,4),
    	type VARCHAR(100)
)`
	_, err := r.conn.Exec(ctx, q)
	return err
}

func (r *repo) FillTable(ctx context.Context, bord Borders, amount int) error {
	batch := &pgx.Batch{}

	q := `INSERT INTO geo_objects (longitude,latitude,type) VALUES ($1,$2,$3)`

	for i := 0; i < amount; i++ {
		lon := generateRandomFloatNumberOnTheSeg(bord.MinX, bord.MaxX)
		lat := generateRandomFloatNumberOnTheSeg(bord.MinY, bord.MaxY)
		objType := typeOfObject[generateRandomIntNumberOnTheSeg(0, len(typeOfObject)-1)]
		batch.Queue(q, lon, lat, objType)
	}

	results := r.conn.SendBatch(ctx, batch)
	return results.Close()
}
