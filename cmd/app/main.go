package main

import (
	"context"
	"github.com/OkDenAl/mbtiles_converter/config"
	"github.com/OkDenAl/mbtiles_converter/pkg/logging"
	pg_geo_table_generator "github.com/OkDenAl/mbtiles_converter/pkg/pg-geo-table-generator"
	"github.com/OkDenAl/mbtiles_converter/pkg/postgres"
)

func main() {
	//dsn := flag.String("dsn", "", "dsn for postgres")
	//filenamePref := flag.String("f", "map_", "prefix for output .mbtiles file")
	//countToConvert := flag.Int("c", 10, "number of rows in the postgres bd to be converted")
	//cfg := config.Config{DB: config.DB{DSN: *dsn, CountToConvert: *countToConvert}, OutFilenamePrefix: *filenamePref}
	log := logging.Init()
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	pool, err := postgres.New(cfg.DSN, 5, log)
	if err != nil {
		log.Fatal(err)
	}
	generator := pg_geo_table_generator.New(pg_geo_table_generator.NewRepo(pool))
	bords := pg_geo_table_generator.Borders{MinX: 37.0471, MaxX: 38.1495, MinY: 55.4652, MaxY: 55.9871}
	err = generator.Generate(context.Background(), bords, 10)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("done")
}
