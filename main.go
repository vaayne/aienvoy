package main

import (
	"embed"
	"log/slog"
	"os"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/readease"
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	_ "github.com/Vaayne/aienvoy/internal/pkg/logger"
	"github.com/Vaayne/aienvoy/internal/ports/httpserver"
	"github.com/Vaayne/aienvoy/internal/ports/tgbot"
	_ "github.com/Vaayne/aienvoy/migrations"

	"github.com/pocketbase/pocketbase/tools/cron"

	tb "gopkg.in/telebot.v3"

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
	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	app := pocketbase.New()

	// migrate DB
	migratecmd.MustRegister(app, app.RootCmd, &migratecmd.Options{
		Automigrate: isGoRun,
	})

	// scheduled jobs
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		scheduler := cron.New()
		// every 5 minutes to run readease job
		if config.GetConfig().ReadEase.TelegramChannel != 0 {
			scheduler.MustAdd("readease", "0 * * * *", func() {
				summaries, err := readease.ReadEasePeriodJob(app)
				if err != nil {
					slog.Error("run period readease job error", "err", err)
				}
				bot := tgbot.DefaultBot(app)
				channel := tb.ChatID(config.GetConfig().ReadEase.TelegramChannel)
				for _, summary := range summaries {
					msg, err := bot.Send(channel, summary)
					if err != nil {
						slog.Error("failed to send readease message to channel", "err", err, "msg", msg)
					}
				}
			})
		}
		scheduler.Start()
		return nil
	})
	// register routes
	registerRoutes(app)
	// start telegram bot readease
	go tgbot.Serve(app)
	if err := app.Start(); err != nil {
		slog.Error("failed to start app", "err", err)
	}
}
