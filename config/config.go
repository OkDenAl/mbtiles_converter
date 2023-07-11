package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		DB                `yaml:"db"`
		OutFilenamePrefix string `yaml:"out_filename_prefix"`
	}
	DB struct {
		DSN            string `yaml:"dsn"`
		CountToConvert int    `yaml:"count_to_convert"`
	}
)

func New() (Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig("./config/config.yml", &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("error while read config: %w", err)
	}
	return cfg, nil
}
