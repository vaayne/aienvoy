package tgbot

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"

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
		Token: token, // Poller:  WebHook,
		// Verbose: false,
	})
	if err != nil {
		slog.Error("Init telegram bot error", "err", err)
		return nil
	}

	return &TeleBot{
		Bot: b,
		app: app,
	}
}

func DefaultBot(app *pocketbase.PocketBase) *TeleBot {
	if bot == nil {
		once.Do(func() {
			bot = New(config.GetConfig().Telegram.Token, app)
		})
	}
	return bot
}

func contextMiddleware(next tb.HandlerFunc) tb.HandlerFunc {
	// nolint:staticcheck
	return func(c tb.Context) error {
		ctx := context.Background()
		ctx = context.WithValue(ctx, config.ContextKeyApp, bot.app)
		ctx = context.WithValue(ctx, config.ContextKeyDao, bot.app.Dao())
		ctx = context.WithValue(ctx, config.ContextKeyUserId, c.Sender().ID)
		ctx = context.WithValue(ctx, config.ContextKeyRequestId, uuid.NewString())
		c.Set(config.ContextKeyContext, ctx)
		return next(c)
	}
}

func registerCommands(b *TeleBot) {
	cmds := []tb.Command{
		{
			Text:        handler.CommandRead,
			Description: "ReadEase to summary article or video using Claude 2",
		},
		{
			Text:        handler.CommandBard,
			Description: "Chat using Google Bard",
		},
		{
			Text:        handler.CommandClaude,
			Description: "Chat using Claude Web",
		},
		{
			Text:        handler.CommandChatGPT35,
			Description: "Chat using ChatGPT 3.5",
		},
		{
			Text:        handler.CommandChatGPT4,
			Description: "Chat using ChatGPT 4",
		},
	}
	if err := b.SetCommands(cmds); err != nil {
		slog.Error("set telegram bot commands error", "err", err)
	} else {
		slog.Info("success set commands")
	}
}

func registerHandlers(b *TeleBot) {
	b.Handle(tb.OnText, handler.OnText)
}

func Serve(app *pocketbase.PocketBase) {
	b := DefaultBot(app)
	b.Use(contextMiddleware)
	registerHandlers(b)
	registerCommands(b)
	slog.Info("Start telegram bot...")
	b.Start()
}
