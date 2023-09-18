package handler

import (
	"fmt"
	"log/slog"

	tb "gopkg.in/telebot.v3"
)

func BardChat(c tb.Context) error {
	slog.Info("bard chat", "text", c.Text())
	return c.Reply(fmt.Sprintf("Bard: %s", c.Text()))
}
