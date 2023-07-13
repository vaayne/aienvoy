package main

import (
	"embed"

	"openai-dashboard/internal/pkg/logger"
	"openai-dashboard/internal/ports/httpserver"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

//go:embed all:web
var staticFiles embed.FS

func registerRoutes(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		httpserver.RegisterRoutes(e.Router, app, staticFiles)
		return nil
	})
}

func main() {
	app := pocketbase.New()
	app.Dao()
	registerRoutes(app)
	if err := app.Start(); err != nil {
		logger.SugaredLogger.Fatalw("failed to start app", "error", err)
	}
}
