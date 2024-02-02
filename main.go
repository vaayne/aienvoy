package main

import (
	"embed"
	"log/slog"
	"os/exec"
	"runtime"

	"github.com/Vaayne/aienvoy/internal/core/midjourney"
	"github.com/Vaayne/aienvoy/internal/core/readease"
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	_ "github.com/Vaayne/aienvoy/internal/pkg/logger"
	"github.com/Vaayne/aienvoy/internal/ports/httpserver"
	"github.com/Vaayne/aienvoy/internal/ports/tgbot"
	_ "github.com/Vaayne/aienvoy/migrations"
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
	"github.com/pocketbase/pocketbase/tools/cron"
	tb "gopkg.in/telebot.v3"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

//go:embed all:web
var staticFiles embed.FS

// RegisterRoutes registers the HTTP routes for the application.
func RegisterRoutes(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		httpserver.RegisterRoutes(e.Router, app, staticFiles)
		return nil
	})
}

// SetScheduledJobs sets up the scheduled jobs for the application.
func SetScheduledJobs(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		scheduler := cron.New()
		// hourly readease job
		if config.GetConfig().ReadEase.TelegramChannel != 0 {
			scheduler.MustAdd("readease", "0 * * * *", func() {
				summaries, err := readease.PeriodJob(app, llm.DefaultGeminiModel)
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
}

// StartTelegramBot starts the Telegram bot for the application.
func StartTelegramBot(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		if config.GetConfig().Telegram.Token != "" {
			go tgbot.Serve(app)
		}
		return nil
	})
}

// StartMidjourneyServer starts the Midjourney server for the application.
func StartMidjourneyServer(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		if config.GetConfig().MidJourney.DiscordBotToken != "" {
			m := midjourney.New(app.Dao())
			m.Client.Serve()
		}
		return nil
	})
}

// OpenBrowser opens the default web browser with the specified URL.
func OpenBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}

	if err := exec.Command(cmd, args...).Start(); err != nil {
		slog.Error("Open browser error", "err", err)
	}
}

func main() {
	app := pocketbase.New()
	// migrate DB
	migratecmd.MustRegister(app, app.RootCmd, &migratecmd.Options{
		Automigrate: false,
	})

	// before serve hooks
	RegisterRoutes(app)
	StartTelegramBot(app)
	StartMidjourneyServer(app)
	// SetScheduledJobs(app)
	// OpenBrowser(config.GetConfig().Service.URL)

	// start app
	if err := app.Start(); err != nil {
		slog.Error("failed to start app", "err", err)
	}
}
