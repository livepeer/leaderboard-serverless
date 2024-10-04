package common

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
)

type ILogger interface {
	Trace(msg string, vars ...interface{})
	Debug(msg string, vars ...interface{})
	Info(msg string, vars ...interface{})
	Warn(msg string, vars ...interface{})
	Error(msg string, vars ...interface{})
	Fatal(msg string, vars ...interface{})
	SetLevel(level string)
}

// Logger is the logger that will be used throughout the application.
// as the application codebase scales, consider placing this in a context.
var Logger ILogger = NewSlogLogger()

type SlogLogger struct {
	log           *slog.Logger
	level         *slog.LevelVar
	internalLevel string
}

func NewSlogLogger() ILogger {
	internalLevel := strings.ToLower(EnvOrDefault("LOG_LEVEL", "info").(string))
	slogLevel, _ := convertToSlogLevel(internalLevel)

	lvl := new(slog.LevelVar)
	lvl.Set(slogLevel)
	sl := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))

	return &SlogLogger{log: sl, level: lvl}
}

func (sl *SlogLogger) Trace(msg string, vars ...interface{}) {
	if sl.internalLevel == "trace" {
		sl.log.Debug(fmt.Sprintf(msg, vars...))
	}
}

func (sl *SlogLogger) Debug(msg string, vars ...interface{}) {
	sl.log.Debug(fmt.Sprintf(msg, vars...))
}

func (sl *SlogLogger) Info(msg string, vars ...interface{}) {
	sl.log.Info(fmt.Sprintf(msg, vars...))
}

func (sl *SlogLogger) Warn(msg string, vars ...interface{}) {
	sl.log.Warn(fmt.Sprintf(msg, vars...))
}

func (sl *SlogLogger) Error(msg string, vars ...interface{}) {
	sl.log.Error(fmt.Sprintf(msg, vars...))
}

func (sl *SlogLogger) Fatal(msg string, vars ...interface{}) {
	sl.log.Error(fmt.Sprintf(msg, vars...))
	log.Fatal(msg)
}

func (sl *SlogLogger) SetLevel(level string) {
	if level == "" {
		level = "INFO"
	}
	lvl, err := convertToSlogLevel(strings.ToLower(level))
	if err != nil {
		log.Printf("Invalid log level: %s.  Will use default of INFO", level)
	}
	sl.log.Info(fmt.Sprintf("Setting log level to %s", lvl.Level().String()))
	sl.level.Set(lvl)
}

func convertToSlogLevel(level string) (slog.Level, error) {
	slogLevel := slog.LevelInfo
	if level != "" {
		switch level {
		case "trace":
			slogLevel = slog.LevelDebug
		case "debug":
			slogLevel = slog.LevelDebug
		case "info":
			slogLevel = slog.LevelInfo
		case "warn":
			slogLevel = slog.LevelWarn
		case "error":
			slogLevel = slog.LevelError
		default:
			log.Printf("Invalid log level: %s.  Will use default of INFO", level)
		}
	}
	return slogLevel, nil
}
