package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/claudeweb"

	"github.com/Vaayne/aienvoy/internal/core/llm/llmclaudeweb"
	tb "gopkg.in/telebot.v3"
)

const claudeModelName = "claude2"

func OnClaudeChat(c tb.Context) error {
	text := strings.TrimSpace(c.Text()[1+len(CommandClaudeWeb):])
	if text == "" {
		text = "hello"
	}
	return askClaude(c, "", text)
}

func askClaude(c tb.Context, id, prompt string) error {
	ctx := c.Get(config.ContextKeyContext).(context.Context)
	claude := llmclaudeweb.New()
	if id == "" {
		cov, err := claude.CreateConversation(prompt[:min(10, len(prompt))])
		if err != nil {
			return c.Reply("Create new claude conversiton error: %s", err)
		}
		id = cov.UUID
	}

	respChan := make(chan *claudeweb.ChatMessageResponse)
	errChan := make(chan error)
	defer close(respChan)
	defer close(errChan)
	msg, err := c.Bot().Send(c.Sender(), "Waiting for response ...")
	if err != nil {
		return fmt.Errorf("chat with ChatGPT err: %v", err)
	}
	go claude.CreateChatMessageStream(id, prompt, respChan, errChan)

	text := ""
	chunk := ""

	for {
		select {
		case resp := <-respChan:
			text, chunk = processResponse(c, ctx, msg, resp.Completion, text, chunk)
		case err := <-errChan:
			newErr := processError(c, ctx, msg, text, err)
			if errors.Is(err, io.EOF) {
				setLLMConversationToCache(LLMCache{
					Model:        claudeModelName,
					Conversation: id,
				})
			}
			return newErr
		case <-ctx.Done():
			return processContextDone(ctx)
		}
	}
}
