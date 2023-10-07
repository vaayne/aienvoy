package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm/llmopenai"
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

const (
	modelNameGPT3Dot5Turbo = openai.GPT3Dot5Turbo
	modelNameGPT4          = openai.GPT4
)

func OnChatGPTChat(c tb.Context, model string) error {
	text := strings.TrimSpace(c.Data())
	if text == "" {
		text = "hello"
	}
	return askChatGPT(c, "", model, text, nil)
}

func OnChatGPT35(c tb.Context) error {
	return OnChatGPTChat(c, modelNameGPT3Dot5Turbo)
}

func OnChatGPT4(c tb.Context) error {
	return OnChatGPTChat(c, modelNameGPT4)
}

func askChatGPT(c tb.Context, id, model, prompt string, messages []openai.ChatCompletionMessage) error {
	if id == "" {
		id = uuid.New().String()
	}
	if messages == nil {
		messages = make([]openai.ChatCompletionMessage, 0)
	}
	llm := llmopenai.New()
	ctx := c.Get(config.ContextKeyContext).(context.Context)

	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		}),
		Stream: false,
	}
	resp, err := llm.Chat(ctx, req)
	if err != nil {
		return c.Reply(fmt.Sprintf("Chat with ChatGPT error: %s", err))
	}
	setLLMConversationToCache(LLMCache{
		Model:        model,
		Conversation: id,
		Messages:     append(messages, resp.Choices[0].Message),
	})
	return c.Send(resp.Choices[0].Message.Content)
}
