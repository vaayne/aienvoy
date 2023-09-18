package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/loghandler"
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

	slog.SetDefault(slog.New(loghandler.NewHandler(handler, config.ContextKeyUserId, config.ContextKeyRequestId)))
}

func IsDebug() bool {
	return strings.ToLower(config.GetConfig().Service.LogLevel) == "debug"
}
