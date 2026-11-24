package config

import "os"

type Config struct {
	Environment string
	HTTPAddr    string
}

func Load() Config {
	cfg := Config{
		Environment: os.Getenv("APP_ENV"),
		HTTPAddr:    os.Getenv("HTTP_ADDR"),
	}
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}
	return cfg
}
