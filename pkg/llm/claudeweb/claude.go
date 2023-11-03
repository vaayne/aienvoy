package claudeweb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llm"
)

// ClaudeWeb is a Claude request client
type ClaudeWeb struct {
	llm.LLM
	Client *Client
}

func New(sessionKey string) *ClaudeWeb {
	cli := NewClient(sessionKey)
	return &ClaudeWeb{
		Client: cli,
		LLM:    llm.LLM{},
	}
}

func (cw *ClaudeWeb) ListModels() []string {
	return ListModels()
}

func ListModels() []string {
	return []string{ModelClaudeWeb}
}

func (cw *ClaudeWeb) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	slog.InfoContext(ctx, "chat with Claude Web start")
	prompt := req.ToPrompt()
	cov, err := cw.Client.CreateConversation(prompt[:min(10, len(prompt))])
	if err != nil {
		slog.ErrorContext(ctx, "chat with Claude Web error", "err", err)
		return llm.ChatCompletionResponse{}, fmt.Errorf("create new claude conversiton error: %w", err)
	}

	resp, err := cw.Client.CreateChatMessage(cov.UUID, prompt)
	if err != nil {
		slog.ErrorContext(ctx, "chat with Claude Web error", "err", err)
		return llm.ChatCompletionResponse{}, fmt.Errorf("chat with claude error: %v", err)
	}
	slog.InfoContext(ctx, "chat with Claude Web success")
	return resp.ToChatCompletionResponse(), nil
}

func (cw *ClaudeWeb) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat with Claude Web stream start")
	prompt := req.ToPrompt()
	cov, err := cw.Client.CreateConversation(prompt[:min(10, len(prompt))])
	if err != nil {
		errChan <- fmt.Errorf("create new claude conversiton error: %w", err)
		return
	}

	messageChan := make(chan *ChatMessageResponse)
	innerErrChan := make(chan error)

	go cw.Client.CreateChatMessageStream(cov.UUID, prompt, messageChan, innerErrChan)
	sb := strings.Builder{}
	for {
		select {
		case resp := <-messageChan:
			sb.WriteString(resp.Completion)
			dataChan <- resp.ToChatCompletionStreamResponse()
		case err := <-innerErrChan:
			if errors.Is(err, io.EOF) {
				slog.InfoContext(ctx, "claude stream done", "cov_id", cov.UUID)
				slog.InfoContext(ctx, "chat with Claude Web stream success")
				errChan <- err
				return
			}
			slog.ErrorContext(ctx, "chat with Claude Web error", "err", err)
			errChan <- err
			return
		}
	}
}
