package logs

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/Alijeyrad/simorq_backend/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New builds a logger from config, supporting multi-output fan-out.
func New(cfg *config.Config) *slog.Logger {
	level := parseLevel(cfg.Logging.Level)
	isDev := strings.EqualFold(cfg.Server.Environment, "development")

	var writers []io.Writer

	// Always write to stdout if enabled or nothing else is configured
	if cfg.Logging.Output.Stdout || (!cfg.Logging.Output.File.Enabled && !cfg.Logging.Output.Loki.Enabled) {
		writers = append(writers, os.Stdout)
	}

	// File output with rotation via lumberjack
	if cfg.Logging.Output.File.Enabled {
		writers = append(writers, &lumberjack.Logger{
			Filename:   cfg.Logging.Output.File.Path,
			MaxSize:    cfg.Logging.Output.File.MaxSizeMB,
			MaxBackups: cfg.Logging.Output.File.MaxBackups,
			MaxAge:     cfg.Logging.Output.File.MaxAgeDays,
			Compress:   cfg.Logging.Output.File.Compress,
		})
	}

	var handlers []slog.Handler

	// Build handler(s) for file/stdout writers
	if len(writers) > 0 {
		w := io.MultiWriter(writers...)
		opts := &slog.HandlerOptions{
			Level:     level,
			AddSource: isDev,
		}
		if strings.EqualFold(cfg.Logging.Format, "json") || !isDev {
			handlers = append(handlers, slog.NewJSONHandler(w, opts))
		} else {
			handlers = append(handlers, slog.NewTextHandler(w, opts))
		}
	}

	// Loki handler via HTTP (no extra dependency — uses slog + HTTP writer)
	if cfg.Logging.Output.Loki.Enabled {
		handlers = append(handlers, newLokiHandler(cfg, level))
	}

	var h slog.Handler
	if len(handlers) == 1 {
		h = handlers[0]
	} else {
		h = &multiHandler{handlers: handlers}
	}

	return slog.New(h).With(
		slog.String("service", cfg.Observability.ServiceName),
		slog.String("version", cfg.Observability.ServiceVersion),
		slog.String("env", cfg.Server.Environment),
	)
}

func Default() *slog.Logger {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
	})
	return slog.New(h).With(slog.String("service", "simorq_backend"))
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
