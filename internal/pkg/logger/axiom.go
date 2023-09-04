package logger

// func initAxiom() {
// 	// Log initialization message
// 	Logger.Sugar().Info("Enabling Axiom logging", "dataset", config.GetConfig().Axiom.Dataset)

// 	// Create Axiom core object
// 	axiomCore, err := adapter.New(
// 		adapter.SetClientOptions(axiom.SetAPITokenConfig(config.GetConfig().Axiom.Token)),
// 		adapter.SetDataset(config.GetConfig().Axiom.Dataset),
// 	)
// 	if err != nil {
// 		// Log fatal error and exit program if there's an issue creating the Axiom core object
// 		Logger.Sugar().Fatalf("Failed to create Axiom core object: %v", err)
// 	}

// 	// Add Axiom core option to logger

// 	Logger = Logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
// 		return zapcore.NewTee(c, axiomCore)
// 	}))

// 	// Start a background goroutine to periodically sync logs to Axiom
// 	go func() {
// 		ticker := time.NewTicker(5 * time.Second)
// 		for range ticker.C {
// 			if err := Logger.Sync(); err != nil && err.Error() != "sync /dev/stderr: invalid argument" {
// 				// Log non-fatal error if there's an issue syncing logs to Axiom
// 				Logger.Sugar().Error("Axiom failed to sync logs", "error", err)
// 			}
// 		}
// 	}()
// }
