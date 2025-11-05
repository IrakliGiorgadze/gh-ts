package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if env == "dev" {
		l = l.Level(zerolog.DebugLevel)
	} else {
		l = l.Level(zerolog.InfoLevel)
	}
	return l
}
