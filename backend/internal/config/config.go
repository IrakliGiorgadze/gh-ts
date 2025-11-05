package config

import "os"

type Config struct {
	Env    string
	Port   string
	DBURL  string
	Origin string // CORS
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func Load() Config {
	return Config{
		Env:    env("APP_ENV", "dev"),
		Port:   env("API_PORT", "8080"),
		DBURL:  env("DB_DSN", "postgres://ticketuser:ticketpass123@localhost:5432/ticketing_db?sslmode=disable"),
		Origin: env("CORS_ORIGIN", "http://localhost:3000"),
	}
}
