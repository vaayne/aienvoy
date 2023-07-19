package logger

import (
	"log"
	"time"

	"aienvoy/internal/pkg/config"
	"aienvoy/internal/pkg/context"

	"go.uber.org/zap/zapcore"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

var (
	SugaredLogger *zap.SugaredLogger
	Logger        *zap.Logger
)

const (
	LevelDebug = "debug"
)

func init() {
	// Parse log level from config
	level, err := zapcore.ParseLevel(config.GetConfig().Service.LogLevel)
	if err != nil {
		// Log fatal error and exit program if there's an issue parsing the log level
		log.Fatalf("Failed to parse log level: %v", err)
	}

	// Create logger based on log level
	if level == zapcore.DebugLevel {
		Logger, _ = zap.NewDevelopment()
	} else {
		Logger, _ = zap.NewProduction()
	}

	// Initialize Axiom logger if token is present in config
	if config.GetConfig().Axiom.Token != "" {
		initAxiom()
	}

	// Assign SugaredLogger and defer sync for later
	SugaredLogger = Logger.Sugar().With(
		"service", config.GetConfig().Service.Name,
		"env", config.GetConfig().Service.Env,
	)
	// nolint:errcheck
	defer Logger.Sync()
}

func initAxiom() {
	// Log initialization message
	Logger.Sugar().Infow("Enabling Axiom logging", "dataset", config.GetConfig().Axiom.Dataset)

	// Create Axiom core object
	axiomCore, err := adapter.New(
		adapter.SetClientOptions(axiom.SetAPITokenConfig(config.GetConfig().Axiom.Token)),
		adapter.SetDataset(config.GetConfig().Axiom.Dataset),
	)
	if err != nil {
		// Log fatal error and exit program if there's an issue creating the Axiom core object
		Logger.Sugar().Fatalf("Failed to create Axiom core object: %v", err)
	}

	// Add Axiom core option to logger
	Logger = Logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, axiomCore)
	}))

	// Start a background goroutine to periodically sync logs to Axiom
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			if err := Logger.Sync(); err != nil && err.Error() != "sync /dev/stderr: invalid argument" {
				// Log non-fatal error if there's an issue syncing logs to Axiom
				Logger.Sugar().Errorw("Axiom failed to sync logs", "error", err)
			}
		}
	}()
}

func GetSugaredLoggerWithContext(ctx context.Context) *zap.SugaredLogger {
	newFields := []interface{}{
		"request_id", ctx.RequestId(),
		"user_id", ctx.UserId(),
	}
	return SugaredLogger.With(newFields...)
}

func GetSugaredLoggerWithEchoContext(ctx echo.Context) *zap.SugaredLogger {
	return GetSugaredLoggerWithContext(context.FromEchoContext(ctx))
}
