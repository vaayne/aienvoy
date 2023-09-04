package logger

import (
	"log/slog"
	"os"
	"strings"

	"aienvoy/internal/pkg/config"
)

func init() {
	var handler slog.Handler
	if IsDebug() {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		})
	}

	slog.SetDefault(slog.New(NewHandler(handler)))
}

func IsDebug() bool {
	return strings.ToLower(config.GetConfig().Service.LogLevel) == "debug"
}
