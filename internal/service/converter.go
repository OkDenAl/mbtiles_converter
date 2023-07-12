package service

import (
	"github.com/OkDenAl/mbtiles_converter/internal/repository/pg"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/sqlite"
)

type Converter interface {
	Convert() error
}

type converter struct {
	pgRepo     pg.Repository
	sqliteRepo sqlite.Repository
}

func NewConverter(pgRepo pg.Repository, sqliteRepo sqlite.Repository) Converter {
	return &converter{pgRepo: pgRepo, sqliteRepo: sqliteRepo}
}

func (c *converter) Convert() error {
	return nil
}
