package logger

import (
	"context"
	"log"
	"time"

	"aienvoy/internal/pkg/config"

	"go.uber.org/zap/zapcore"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
	"github.com/axiomhq/axiom-go/axiom"
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

// SugaredLoggerWithContext returns a new sugared logger with additional fields
// extracted from the given context using the provided keys.
func SugaredLoggerWithContext(ctx context.Context, keys ...string) *zap.SugaredLogger {
	// Pre-allocate a slice of interfaces for the new fields to add to the logger.
	newFields := make([]interface{}, 0, len(keys)*2)

	// Iterate through the keys and add the corresponding values from the context to the new fields.
	for _, key := range keys {
		newFields = append(newFields, key, ctx.Value(key))
	}

	// Add the new fields to the existing SugaredLogger and return the result.
	return SugaredLogger.With(newFields...)
}
