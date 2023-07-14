package main

import (
	"context"
	"database/sql"
	"github.com/OkDenAl/mbtiles_converter/config"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/pg"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/sqlite"
	"github.com/OkDenAl/mbtiles_converter/internal/service"
	"github.com/OkDenAl/mbtiles_converter/pkg/logging"
	"github.com/OkDenAl/mbtiles_converter/pkg/pg_geo_table_generator"
	"github.com/OkDenAl/mbtiles_converter/pkg/postgres"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log := logging.Init()
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("connecting to postgres...")
	pool, err := postgres.New(cfg.DSN, 5, log)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	log.Info("successfully connected")

	if cfg.NeedToGenerateData {
		err = pg_geo_table_generator.Run(pool)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Info("connecting to sqlite...")
	//sqliteFilename := fmt.Sprintf("mbtiles/%s_%s.mbtiles", cfg.OutFilenamePrefix, time.Now().String())
	db, err := sql.Open("sqlite3", "mbtiles/test.mbtiles")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Info("successfully connected")

	log.Info("converting data...")
	converter := service.NewConverter(pg.NewRepo(pool), sqlite.NewRepo(db))
	err = converter.Convert(context.Background(), cfg.ConverterOpts)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("done")
}
