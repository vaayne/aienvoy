package bard

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/Vaayne/aienvoy/pkg/llm"
)

const ModelBard = "bard"

func ListModels() []string {
	return []string{ModelBard}
}

type Bard struct {
	*Client
	*llm.LLM
}

func New(token string, opts ...ClientOption) (*Bard, error) {
	client, err := NewClient(token, opts...)
	if err != nil {
		return nil, err
	}
	return &Bard{
		Client: client,
		LLM:    &llm.LLM{},
	}, nil
}

func (c *Bard) ListModels() []string {
	return ListModels()
}

func (c *Bard) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	slog.InfoContext(ctx, "chat with Google Bard start")
	prompt := req.ToPrompt()
	resp, err := c.Ask(prompt, "", "", "", 0)
	if err != nil {
		slog.ErrorContext(ctx, "chat with Google Bard error", "err", err)
		return llm.ChatCompletionResponse{}, fmt.Errorf("bard got an error, %w", err)
	}
	res := resp.ToChatCompletionResponse()
	slog.InfoContext(ctx, "chat with Google Bard success")
	return res, nil
}

func (c *Bard) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat with Google Bard stream start")
	prompt := req.ToPrompt()
	resp, err := c.Ask(prompt, "", "", "", 0)
	if err != nil {
		errChan <- fmt.Errorf("bard got an error, %w", err)
		slog.ErrorContext(ctx, "chat with Google Bard stream error", "err", err)
		return
	}
	res := resp.ToChatCompletionStreamResponse()
	slog.InfoContext(ctx, "chat with Google Bard stream success")
	dataChan <- res
	errChan <- io.EOF
}
