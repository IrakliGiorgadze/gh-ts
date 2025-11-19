package config

import (
	"fmt"
	"os"
)

type Config struct {
	Env           string
	Port          string
	DBURL         string
	Origin        string // CORS
	SessionSecret string
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envRequired(k string) (string, error) {
	v := os.Getenv(k)
	if v == "" {
		return "", fmt.Errorf("required environment variable %s is not set", k)
	}
	return v, nil
}

func Load() (Config, error) {
	dbURL, err := envRequired("DB_DSN")
	if err != nil {
		return Config{}, err
	}

	sessionSecret, err := envRequired("SESSION_SECRET")
	if err != nil {
		return Config{}, err
	}

	return Config{
		Env:           env("APP_ENV", "dev"),
		Port:          env("API_PORT", "8080"),
		DBURL:         dbURL,
		Origin:        env("CORS_ORIGIN", "http://localhost:3000"),
		SessionSecret: sessionSecret,
	}, nil
}
