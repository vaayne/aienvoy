package tgbot

import (
	"log/slog"
	"sync"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/internal/ports/tgbot/handler"

	"github.com/pocketbase/pocketbase"
	tb "gopkg.in/telebot.v3"
)

type TeleBot struct {
	*tb.Bot
	// app is use for db usage
	app *pocketbase.PocketBase
}

var (
	bot  *TeleBot
	once sync.Once
)

func New(token string, app *pocketbase.PocketBase) *TeleBot {
	b, err := tb.NewBot(tb.Settings{
		Token: token,
		// Poller:  WebHook,
		// Verbose: false,
	})
	if err != nil {
		slog.Error("Init telebot error", "err", err)
		return nil
	}

	teleBot := &TeleBot{
		Bot: b,
		app: app,
	}
	// teleBot.registerHandlers()

	return teleBot
}

func DefaultBot(app *pocketbase.PocketBase) *TeleBot {
	if bot == nil {
		once.Do(func() {
			bot = New(config.GetConfig().Telegram.Token, app)
		})
	}
	return bot
}

func appMiddleware(next tb.HandlerFunc) tb.HandlerFunc {
	return func(c tb.Context) error {
		c.Set("app", bot.app)
		return next(c) // continue execution chain
	}
}

func registerCommands(b *TeleBot) {
	cmds := []tb.Command{
		{
			Text:        handler.CommandRead,
			Description: "ReadEase to summary artilcle or vedio using Claude 2",
		},
		{
			Text:        handler.CommandBard,
			Description: "Chat using Google Bard",
		},
		{
			Text:        handler.CommandClaude,
			Description: "Chat using Claude Web",
		},
	}
	if err := b.SetCommands(cmds); err != nil {
		slog.Error("set telegram bot commands error", "err", err)
	}
}

func registerHandlers(b *TeleBot) {
	b.Handle(tb.OnText, handler.OnText)
	// b.Handle(commandRead, handler.OnReadEase)
	// b.Handle(commandBard, handler.BardChat)
	// b.Handle(commandClaude, handler.ClaudeChat)
}

func Serve(app *pocketbase.PocketBase) {
	b := DefaultBot(app)
	b.Use(appMiddleware)
	registerHandlers(b)
	registerCommands(b)
	slog.Info("Start telegram bot...")
	b.Start()
}
