package handler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/llms/llm"

	"github.com/Vaayne/aienvoy/internal/core/readease"
	"github.com/pocketbase/pocketbase"
	tb "gopkg.in/telebot.v3"
)

var InvalidURLError = errors.New("invalid url")

func OnReadEase(c tb.Context) error {
	urlStr := strings.TrimSpace(c.Data())
	ctx := c.Get(config.ContextKeyContext).(context.Context)
	ctx, cancel := context.WithTimeout(ctx, 60*10*time.Second)
	defer cancel()
	_, err := url.ParseRequestURI(urlStr)
	if err != nil || !strings.HasPrefix(urlStr, "http") {
		return c.Reply(fmt.Sprintf("invalid url %s, please check and try again", urlStr))
	}

	msg, err := c.Bot().Send(c.Sender(), "please wait a moment, I am reading the article...")
	if err != nil {
		return fmt.Errorf("summary article err: %v", err)
	}

	reader := readease.NewReader(ctx.Value(config.ContextKeyApp).(*pocketbase.PocketBase))

	respChan := make(chan llm.ChatCompletionStreamResponse)
	errChan := make(chan error)
	defer close(respChan)
	defer close(errChan)

	go reader.ReadStream(ctx, urlStr, "gemini-pro", respChan, errChan)

	text := ""
	chunk := ""

	for {
		select {
		case resp := <-respChan:
			text, chunk = processResponse(c, ctx, msg, resp.Choices[0].Delta.Content, text, chunk)
		case err := <-errChan:
			return processError(c, ctx, msg, text, err)
		case <-ctx.Done():
			return processContextDone(ctx)
		}
	}
}
