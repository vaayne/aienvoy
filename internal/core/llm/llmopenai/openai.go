package llmopenai

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// OpenAI is a client for AI services
type OpenAI struct{}

func New() *OpenAI {
	return &OpenAI{}
}

func (s *OpenAI) GetModels(ctx context.Context) (openai.ModelsList, error) {
	return getClientByModel(openai.GPT3Dot5Turbo).ListModels(ctx)
}

func (s *OpenAI) Chat(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	resp, err := getClientByModel(req.Model).CreateChatCompletion(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err.Error())
		return openai.ChatCompletionResponse{}, err
	}

	_ = saveUsage(ctx, req.Model, resp.Usage.TotalTokens)

	return resp, nil
}

func (s *OpenAI) ChatStream(ctx context.Context, req openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error) {
	stream, err := getClientByModel(req.Model).CreateChatCompletionStream(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err.Error())
		errChan <- err
		return
	}

	// log chat stream response
	sb := &strings.Builder{}
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				tk := NewTikToken(req.Model)
				totalTokens := tk.Encode(sb.String())
				_ = saveUsage(ctx, req.Model, totalTokens)
				errChan <- err
				return
			}
			slog.ErrorContext(ctx, "chat stream receive error", "err", err.Error())
			errChan <- err
			return
		}
		if len(resp.Choices) > 0 {
			sb.WriteString(resp.Choices[0].Delta.Content)
			dataChan <- resp
		}
	}
}

func (s *OpenAI) CreateEmbeddings(ctx context.Context, req openai.EmbeddingRequest) (openai.EmbeddingResponse, error) {
	resp, err := getClientByModel(req.Model.String()).CreateEmbeddings(ctx, req)
	return resp, err
}
