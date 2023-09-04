package logger

import (
	"log/slog"
	"os"
	"strings"

	"aienvoy/internal/pkg/config"
)

func init() {
	var handler slog.Handler
	if strings.ToLower(config.GetConfig().Service.LogLevel) == "debug" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		})
	}

	slog.SetDefault(slog.New(NewHandler(handler)))
}
