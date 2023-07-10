package main

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"log"
)

func registerRoutes(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Static("/", "web")
		return nil
	})
}

func main() {
	app := pocketbase.New()
	registerRoutes(app)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
