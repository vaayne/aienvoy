package handler

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Vaayne/aienvoy/internal/core/llmdao"
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/llm"
	llmservice "github.com/Vaayne/aienvoy/pkg/llm/service"
	"github.com/pocketbase/pocketbase/daos"
	tb "gopkg.in/telebot.v3"
)

func onLLMChat(c tb.Context, conversationId, model, prompt string) error {
	ctx := c.Get(config.ContextKeyContext).(context.Context)
	svc, err := llmservice.New(model, llmdao.New(ctx.Value(config.ContextKeyDao).(*daos.Dao)))
	if err != nil {
		return fmt.Errorf("init llm service err: %v", err)
	}
	if conversationId == "" {
		cov, err := svc.CreateConversation(ctx, "")
		if err != nil {
			return fmt.Errorf("create conversation err: %v", err)
		}
		conversationId = cov.Id
	}
	req := llm.ChatCompletionRequest{
		Model: model,
		Messages: []llm.ChatCompletionMessage{
			{
				Role:    llm.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Stream: true,
	}

	respChan := make(chan llm.ChatCompletionStreamResponse)
	defer close(respChan)
	errChan := make(chan error)
	defer close(errChan)
	msg, err := c.Bot().Send(c.Sender(), "Waiting for response ...")
	if err != nil {
		return fmt.Errorf("chat with ChatGPT err: %v", err)
	}
	go svc.CreateMessageStream(ctx, conversationId, req, respChan, errChan)
	text := ""
	chunk := ""

	for {
		select {
		case resp := <-respChan:
			text, chunk = processResponse(c, ctx, msg, resp.Choices[0].Delta.Content, text, chunk)
		case err := <-errChan:
			newErr := processError(c, ctx, msg, text, err)
			if errors.Is(err, io.EOF) {
				setLLMConversationToCache(LLMCache{
					Model:          model,
					ConversationId: conversationId,
				})
			}
			return newErr
		case <-ctx.Done():
			return processContextDone(ctx)
		}
	}
}
