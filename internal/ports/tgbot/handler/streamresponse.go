package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	tb "gopkg.in/telebot.v3"
)

func processResponse(c tb.Context, ctx context.Context, msg *tb.Message, textChan, text, chunk string) (string, string) {
	newText := text + textChan
	newChunk := chunk + textChan
	if strings.TrimSpace(chunk) == "" {
		return newText, newChunk
	}
	if len(newChunk) >= 200 {
		// slog.DebugContext(ctx, "response with text", "text", text)
		//nolint
		newMsg, err := c.Bot().Edit(msg, newText)
		if err != nil {
			slog.WarnContext(ctx, "telegram bot edit msg err", "err", err)
		} else {
			//nolint
			msg = newMsg
		}
		newChunk = ""
	}
	return newText, newChunk
}

func processError(c tb.Context, ctx context.Context, msg *tb.Message, text string, err error) error {
	if errors.Is(err, InvalidURLError) {
		return c.Reply("invalid url, please check and try again")
	} else if errors.Is(err, io.EOF) {
		// send last message
		if _, err := c.Bot().Edit(msg, text); err != nil {
			slog.WarnContext(ctx, "telegram send last msg err", "err", err)
			return err
		}
		return nil
	}
	if _, err = c.Bot().Edit(msg, err.Error()); err != nil {
		slog.ErrorContext(ctx, "telegram bot edit msg err", "err", err, "text", text)
	}
	return fmt.Errorf("telegram bot process err: %v", err)
}

func processContextDone(ctx context.Context) error {
	slog.ErrorContext(ctx, "telegram bot stream response timeout", "err", ctx.Err())
	return fmt.Errorf("telegram bot processing timeout, please wait a moment and try again")
}
