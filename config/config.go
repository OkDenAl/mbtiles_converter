package config

import (
	"encoding/json"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type (
	Config struct {
		DB                 `yaml:"db"`
		Metadata           `yaml:"metadata"`
		ConverterOpts      `yaml:"converter_opts"`
		OutFilenamePrefix  string `yaml:"out_filename_prefix"`
		NeedToGenerateData bool   `yaml:"need_to_generate_data"`
		Logger             json.RawMessage
	}
	DB struct {
		DSN       string `yaml:"dsn"`
		TableName string `yaml:"table_name"`
		RowsNames string `yaml:"rows_names"`
	}
	Metadata struct {
		Name    string `yaml:"name"`
		Bounds  string `yaml:"bounds"`
		Center  string `yaml:"center"`
		MinZoom int    `yaml:"min_zoom"`
		MaxZoom int    `yaml:"max_zoom"`
	}
	ConverterOpts struct {
		BatchSize    int `yaml:"batch_size"`
		ConvertLimit int `yaml:"convert_limit"`
		StartZoom    int `yaml:"start_zoom"`
		EndZoom      int `yaml:"end_zoom"`
	}
)

func New() (Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig("./config/config.yml", &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("cleanenv.ReadConfig: %w", err)
	}
	err = cfg.loadJSON("./config/logger.json")
	if err != nil {
		return Config{}, fmt.Errorf("cfg.loadJSON: %w", err)
	}
	return cfg, nil
}

func (cfg *Config) loadJSON(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}
	if err = json.Unmarshal(bytes, cfg); err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}
	return nil
}
