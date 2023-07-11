package main

import (
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/config"
	"github.com/OkDenAl/mbtiles_converter/pkg/logging"
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
	fmt.Println(cfg)
}
