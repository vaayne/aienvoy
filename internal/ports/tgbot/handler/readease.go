package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"aienvoy/internal/core/readease"
	"aienvoy/internal/pkg/logger"
	"aienvoy/pkg/claudeweb"

	"github.com/pocketbase/pocketbase"
	tb "gopkg.in/telebot.v3"
)

var InvalidURLError = errors.New("Inavlid url")

func OnText(c tb.Context) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 60*10*time.Second)
	defer cancel()
	_, err := url.ParseRequestURI(c.Text())
	if err != nil || !strings.HasPrefix(c.Text(), "http") {
		return c.Reply(fmt.Sprintf("invalid url %s, please check and try again", c.Text()))
	}

	msg, err := c.Bot().Send(c.Sender(), "please wait a moment, I am reading the article...")
	if err != nil {
		return fmt.Errorf("summary article err: %v", err)
	}

	reader := readease.NewReader(c.Get("app").(*pocketbase.PocketBase))

	respChan := make(chan *claudeweb.ChatMessageResponse)
	errChan := make(chan error)
	defer close(respChan)
	defer close(errChan)

	go reader.ReadStream(ctx, c.Text(), respChan, errChan)

	text := ""
	chunk := ""

	for {
		select {
		case resp := <-respChan:
			text += resp.Completion
			chunk += resp.Completion
			if strings.TrimSpace(chunk) == "" {
				continue
			}
			if len(chunk) > 200 {
				// logger.SugaredLogger.Debugw("response with text", "text", text)
				newMsg, err := c.Bot().Edit(msg, text)
				if err != nil {
					logger.SugaredLogger.Warnw("onText edit msg err", "err", err)
				} else {
					msg = newMsg
				}
				chunk = ""
			}
		case err := <-errChan:
			if errors.Is(err, InvalidURLError) {
				return c.Reply("invalid url, please check and try again")
			} else if errors.Is(err, io.EOF) {

				// send last message
				if _, err := c.Bot().Edit(msg, text); err != nil {
					logger.SugaredLogger.Errorw("onText edit msg err", "err", err)
					return err
				}
				return nil
			}
			if _, err = c.Bot().Edit(msg, err.Error()); err != nil {
				logger.SugaredLogger.Errorw("OnText edit msg err", "err", err, "text", text)
			}
			return fmt.Errorf("summary article err: %v", err)
		case <-ctx.Done():
			logger.SugaredLogger.Errorw("OnText timeout", "err", ctx.Err())
			return fmt.Errorf("summary article timeout, please wait a moment and try again")
		}
	}
}
