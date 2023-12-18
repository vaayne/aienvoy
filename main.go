package main

import (
	"embed"
	"log/slog"

	"github.com/Vaayne/aienvoy/internal/core/midjourney"
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

// func openBrowser(url string) error {
// 	var cmd string
// 	var args []string

// 	switch runtime.GOOS {
// 	case "windows":
// 		cmd = "cmd"
// 		args = []string{"/c", "start", url}
// 	case "darwin":
// 		cmd = "open"
// 		args = []string{url}
// 	default: // "linux", "freebsd", "openbsd", "netbsd"
// 		cmd = "xdg-open"
// 		args = []string{url}
// 	}

// 	return exec.Command(cmd, args...).Start()
// }

func main() {
	app := pocketbase.New()
	// migrate DB
	migratecmd.MustRegister(app, app.RootCmd, &migratecmd.Options{
		Automigrate: false,
	})

	// scheduled jobs
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		scheduler := cron.New()
		// hourly readease job
		if config.GetConfig().ReadEase.TelegramChannel != 0 {
			scheduler.MustAdd("readease", "0 * * * *", func() {
				summaries, err := readease.PeriodJob(app, "gemini-pro")
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

	// after bootstrap start other service
	app.OnAfterBootstrap().Add(func(e *core.BootstrapEvent) error {
		// start telegram bot
		go tgbot.Serve(app)
		// start midjourney bot
		go func() {
			m := midjourney.New(app.Dao())
			m.Client.Serve()
		}()
		return nil
	})
	// go func() {
	// 	if err := openBrowser(config.GetConfig().Service.URL); err != nil {
	// 		slog.Error("Open browser error", "err", err)
	// 	}
	// }()
	if err := app.Start(); err != nil {
		slog.Error("failed to start app", "err", err)
	}
}
