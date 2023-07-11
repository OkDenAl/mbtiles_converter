package pg_geo_table_generator

import (
	"context"
)

const maxAvailableAmount = 1000

type Borders struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
}

type GeoDBGenerator interface {
	Generate(ctx context.Context, bord Borders, amount int) error
}

type geoGenerator struct {
	repo Repository
}

func New(repo Repository) GeoDBGenerator {
	return &geoGenerator{repo: repo}
}

func (g *geoGenerator) Generate(ctx context.Context, bord Borders, amount int) error {
	err := g.repo.CreateTable(ctx)
	if err != nil {
		return err
	}
	if amount > maxAvailableAmount {
		amount = maxAvailableAmount
	}
	return g.repo.FillTable(ctx, bord, amount)
}
