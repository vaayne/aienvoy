package main

import (
	"embed"

	"aienvoy/internal/pkg/logger"
	"aienvoy/internal/ports/httpserver"
	"aienvoy/internal/ports/tgbot"

	_ "aienvoy/migrations"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
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

	go tgbot.Serve(app)

	migratecmd.MustRegister(app, app.RootCmd, &migratecmd.Options{
		Automigrate: false,
	})

	registerRoutes(app)
	if err := app.Start(); err != nil {
		logger.SugaredLogger.Fatalw("failed to start app", "error", err)
	}
}
