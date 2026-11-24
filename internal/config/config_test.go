package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_ADDR", "127.0.0.1:18080")

	cfg := Load()
	if cfg.Environment != "test" || cfg.HTTPAddr != "127.0.0.1:18080" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
