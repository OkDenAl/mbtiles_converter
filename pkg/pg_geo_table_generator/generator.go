package pg_geo_table_generator

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

func NewGenerator(repo Repository) GeoDBGenerator {
	return &geoGenerator{repo: repo}
}

func (g *geoGenerator) Generate(ctx context.Context, bord Borders, amount int) error {
	err := g.repo.CreateTable(ctx)
	if err != nil {
		return fmt.Errorf("g.repo.CreateTable: %w", err)
	}
	err = g.repo.FillTable(ctx, bord, amount)
	if err != nil {
		return fmt.Errorf("g.repo.FillTable: %w", err)
	}
	return nil
}

func Run(pool *pgxpool.Pool) error {
	generator := NewGenerator(NewRepo(pool))
	err := generator.Generate(context.Background(), MoscowSquareBorders, 10000)
	if err != nil {
		return fmt.Errorf("generator.Generate: %w", err)
	}
	return nil
}
