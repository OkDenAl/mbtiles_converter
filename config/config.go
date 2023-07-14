package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		DB                 `yaml:"db"`
		Metadata           `yaml:"metadata"`
		ConverterOpts      `yaml:"converter_opts"`
		OutFilenamePrefix  string `yaml:"out_filename_prefix"`
		NeedToGenerateData bool   `yaml:"need_to_generate_data"`
	}
	DB struct {
		DSN string `yaml:"dsn"`
	}
	Metadata struct {
		Name    string `yaml:"name"`
		Bounds  string `yaml:"bounds"`
		Center  string `yaml:"center"`
		Type    string `yaml:"type"`
		MinZoom int    `yaml:"min_zoom"`
		MaxZoom int    `yaml:"max_zoom"`
		Json    string `yaml:"json"`
	}
	ConverterOpts struct {
		QantityToConvert int `yaml:"count_to_convert"`
		StartZoom        int `yaml:"start_zoom"`
		EndZoom          int `yaml:"end_zoom"`
		ConverterMode    int `yaml:"converter_mode"`
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
