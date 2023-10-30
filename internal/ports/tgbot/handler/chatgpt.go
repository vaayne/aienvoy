package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/Vaayne/aienvoy/internal/core/llm"
	"github.com/Vaayne/aienvoy/internal/core/llm/llmclaude"
	"io"
	"strings"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

func OnChatGPT35(c tb.Context) error {
	return onLLMChat(c, openai.GPT3Dot5Turbo)
}

func OnChatGPT4(c tb.Context) error {
	return onLLMChat(c, openai.GPT4)
}

func OnClaudeV2(c tb.Context) error {
	return onLLMChat(c, llmclaude.ModelClaudeV2)
}

func OnClaudeV1Dot3(c tb.Context) error {
	return onLLMChat(c, llmclaude.ModelClaudeV1Dot3)
}

func OnClaudeInstant(c tb.Context) error {
	return onLLMChat(c, llmclaude.ModelClaudeInstantV1Dot2)
}

func onLLMChat(c tb.Context, model string) error {
	text := strings.TrimSpace(c.Text()[5:])
	if text == "" {
		text = "hello"
	}
	return askLLM(c, "", model, text, nil)
}

func askLLM(c tb.Context, id, model, prompt string, messages []openai.ChatCompletionMessage) error {
	if id == "" {
		id = uuid.New().String()
	}
	if messages == nil {
		messages = make([]openai.ChatCompletionMessage, 0)
	}
	llmSvc := llm.New(model)
	ctx := c.Get(config.ContextKeyContext).(context.Context)

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}

	respChan := make(chan openai.ChatCompletionStreamResponse)
	defer close(respChan)
	errChan := make(chan error)
	defer close(errChan)
	msg, err := c.Bot().Send(c.Sender(), "Waiting for response ...")
	if err != nil {
		return fmt.Errorf("chat with ChatGPT err: %v", err)
	}
	go llmSvc.CreateChatCompletionStream(ctx, &req, respChan, errChan)
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
					Model:        model,
					Conversation: id,
					Messages: append(messages, openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: text,
					}),
				})
			}
			return newErr
		case <-ctx.Done():
			return processContextDone(ctx)
		}
	}
}
