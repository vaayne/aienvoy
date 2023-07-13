package main

import (
	"log"

	"openai-dashboard/internal/ports/httpserver"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func registerRoutes(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		httpserver.RegisterRoutes(e.Router, app)
		return nil
	})
}

func main() {
	app := pocketbase.New()
	app.Dao()
	registerRoutes(app)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
