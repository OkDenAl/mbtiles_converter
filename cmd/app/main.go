package main

import (
	"database/sql"
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/config"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/pg"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/sqlite"
	"github.com/OkDenAl/mbtiles_converter/internal/service"
	"github.com/OkDenAl/mbtiles_converter/pkg/logging"
	"github.com/OkDenAl/mbtiles_converter/pkg/pg_geo_table_generator"
	"github.com/OkDenAl/mbtiles_converter/pkg/postgres"
	"time"
)

func main() {
	log := logging.Init()
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	pool, err := postgres.New(cfg.DSN, 5, log)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if cfg.NeedToGenerateData {
		err = pg_geo_table_generator.Run(pool)
		if err != nil {
			log.Fatal(err)
		}
	}
	sqliteFilename := fmt.Sprintf("%s_%s.mbtiles", cfg.OutFilenamePrefix, time.Now().String())
	db, err := sql.Open("sqlite3", sqliteFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	converter := service.NewConverter(pg.NewRepo(pool), sqlite.NewRepo(db))
	log.Info("done")
}
