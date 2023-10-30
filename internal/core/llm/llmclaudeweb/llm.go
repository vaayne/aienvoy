package llmclaudeweb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm/dto"
	"github.com/Vaayne/aienvoy/internal/core/llm/usage"
	"github.com/Vaayne/aienvoy/pkg/claudeweb"
	"github.com/sashabaranov/go-openai"
)

const ModelClaudeWeb = "claude-2"

func (cw *ClaudeWeb) CreateChatCompletion(ctx context.Context, req *openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	slog.InfoContext(ctx, "chat with Claude Web start")
	prompt := buildPrompt(req)
	cov, err := cw.CreateConversation(prompt[:min(10, len(prompt))])
	if err != nil {
		slog.ErrorContext(ctx, "chat with Claude Web error", "err", err)
		return openai.ChatCompletionResponse{}, fmt.Errorf("create new claude conversiton error: %w", err)
	}

	resp, err := cw.CreateChatMessage(cov.UUID, prompt)
	if err != nil {
		slog.ErrorContext(ctx, "chat with Claude Web error", "err", err)
		return openai.ChatCompletionResponse{}, fmt.Errorf("chat with claude error: %v", err)
	}
	cr := ClaudeResponse{ChatMessageResponse: *resp}
	_ = usage.SaveFromText(ctx, req.Model, cr.Completion)
	slog.InfoContext(ctx, "chat with Claude Web success")
	return cr.ToOpenAIChatCompletionResponse(), nil
}

func (cw *ClaudeWeb) CreateChatCompletionStream(ctx context.Context, req *openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat with Claude Web stream start")
	prompt := buildPrompt(req)
	cov, err := cw.CreateConversation(prompt[:min(10, len(prompt))])
	if err != nil {
		errChan <- fmt.Errorf("create new claude conversiton error: %w", err)
		return
	}

	messageChan := make(chan *claudeweb.ChatMessageResponse)
	innerErrChan := make(chan error)

	go cw.CreateChatMessageStream(cov.UUID, prompt, messageChan, innerErrChan)
	sb := strings.Builder{}
	for {
		select {
		case resp := <-messageChan:
			sb.WriteString(resp.Completion)
			data := ClaudeResponse{ChatMessageResponse: *resp}
			dataChan <- data.ToOpenAIChatCompletionStreamResponse()
		case err := <-innerErrChan:
			if errors.Is(err, io.EOF) {
				slog.InfoContext(ctx, "claude stream done", "cov_id", cov.UUID)
				_ = usage.SaveFromText(ctx, req.Model, sb.String())
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

func (cw *ClaudeWeb) CreateCompletion(ctx context.Context, req *openai.CompletionRequest) (openai.CompletionResponse, error) {
	chatReq := dto.CompletionRequestToChatCompletionRequest(*req)
	resp, err := cw.CreateChatCompletion(ctx, &chatReq)
	if err != nil {
		return openai.CompletionResponse{}, err
	}
	return dto.ChatCompletionResponseToCompletionResponse(resp), nil
}

func (cw *ClaudeWeb) CreateCompletionStream(ctx context.Context, req *openai.CompletionRequest, dataChan chan openai.CompletionResponse, errChan chan error) {
	chatReq := dto.CompletionRequestToChatCompletionRequest(*req)

	respChan := make(chan openai.ChatCompletionStreamResponse)
	innerErrorChan := make(chan error)

	go cw.CreateChatCompletionStream(ctx, &chatReq, respChan, innerErrorChan)

	for {
		select {
		case resp := <-respChan:
			data := dto.ChatCompletionStreamResponseToCompletionResponse(resp)
			dataChan <- data
		case err := <-innerErrorChan:
			errChan <- err
			return
		}
	}
}
