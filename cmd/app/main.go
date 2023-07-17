package main

import (
	"context"
	"github.com/OkDenAl/mbtiles_converter/config"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/pg"
	"github.com/OkDenAl/mbtiles_converter/internal/repository/sqliterepo"
	"github.com/OkDenAl/mbtiles_converter/internal/service"
	"github.com/OkDenAl/mbtiles_converter/pkg/logging"
	"github.com/OkDenAl/mbtiles_converter/pkg/pg_geo_table_generator"
	"github.com/OkDenAl/mbtiles_converter/pkg/postgres"
	"github.com/OkDenAl/mbtiles_converter/pkg/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

func main() {
	log := logging.Init()
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	log.Info(cfg)

	log.Info("connecting to postgres...")
	pgPool, err := postgres.New(cfg.DSN, 5, log)
	if err != nil {
		log.Fatal(err)
	}
	defer pgPool.Close()
	log.Info("successfully connected")

	if cfg.NeedToGenerateData {
		t := time.Now()
		err = pg_geo_table_generator.Run(pgPool)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(time.Since(t))
	}

	log.Info("connecting to sqlite...")
	//sqliteFilename := fmt.Sprintf("mbtiles/%s_%s.mbtiles", cfg.OutFilenamePrefix, time.Now().String())
	sqlitePool, err := sqlite.New("mbtiles/test.mbtiles", 5, log)
	log.Info("successfully connected")

	log.Info("converting data...")
	t := time.Now()
	converter := service.NewConverter(pg.NewRepo(pgPool), sqliterepo.NewRepo(sqlitePool))
	err = converter.Convert(context.Background(), cfg.ConverterOpts)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(time.Since(t))
	log.Info("done")
}
