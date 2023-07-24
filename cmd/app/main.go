package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/config"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/pg"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/sqliterepo"
	"github.com/OkDenAl/mbtiles_converter/internal/service"
	"github.com/OkDenAl/mbtiles_converter/pkg/logging"
	"github.com/OkDenAl/mbtiles_converter/pkg/pg_geo_table_generator"
	"github.com/OkDenAl/mbtiles_converter/pkg/postgres"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"time"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic(fmt.Errorf("unable to load config: %w", err))
	}
	log, err := logging.New(cfg.Logger)
	if err != nil {
		panic(fmt.Errorf("unable to configure logger: %w", err))
	}

	log.Info("connecting to postgres...")
	pgPool, err := postgres.New(cfg.DSN, 5, log)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to connect to postgres: %w", err))
	}
	defer pgPool.Close()
	log.Info("successfully connected")

	if cfg.NeedToGenerateData {
		err = pg_geo_table_generator.Run(pgPool)
		if err != nil {
			log.Fatal(fmt.Errorf("unable to generate table data for postgres: %w", err))
		}
	}

	log.Info("connecting to sqlite...")
	sqliteFilename := fmt.Sprintf("mbtiles/%s_%s.mbtiles", cfg.OutFilenamePrefix,
		strings.ReplaceAll(time.Now().Format(time.DateTime), ":", "-"))
	db, err := sql.Open("sqlite3", sqliteFilename)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to connect to sqlite: %w", err))
	}
	defer db.Close()
	log.Info("successfully connected")

	log.Info("converting data...")
	converter := service.NewConverter(pg.NewRepo(pgPool), sqliterepo.NewRepo(db))
	err = converter.Convert(context.Background(), cfg.ConverterOpts, cfg.Metadata)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to convert data: %w", err))
	}
	log.Info("successfully converted")
}
