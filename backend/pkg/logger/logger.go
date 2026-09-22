package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init(level string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	lvl := zerolog.InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		lvl = zerolog.DebugLevel
	case "error":
		lvl = zerolog.ErrorLevel
	case "info":
		lvl = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(lvl)
	Log = zerolog.New(os.Stdout).With().Timestamp().Logger()
}
