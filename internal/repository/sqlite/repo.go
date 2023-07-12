package sqlite

import (
	"context"
	"database/sql"
)

type Repository interface {
	AddPoint(ctx context.Context, startZoom, endZoom int) error
}

type repo struct {
	conn *sql.DB
}

func NewRepo(conn *sql.DB) Repository {
	return &repo{conn: conn}
}

func (r *repo) AddPoint(ctx context.Context, startZoom, endZoom int) error {
	return nil
}
