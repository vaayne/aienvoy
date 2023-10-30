package llmopenai

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm/usage"

	"github.com/sashabaranov/go-openai"
)

// OpenAI is a client for AI services
type OpenAI struct{}

func New() *OpenAI {
	return &OpenAI{}
}

func (s *OpenAI) CreateChatCompletion(ctx context.Context, req *openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	resp, err := getClientByModel(ctx, req.Model).CreateChatCompletion(ctx, *req)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err.Error())
		return openai.ChatCompletionResponse{}, err
	}
	_ = usage.Save(ctx, req.Model, resp.Usage.TotalTokens)
	return resp, nil
}

func (s *OpenAI) CreateChatCompletionStream(ctx context.Context, req *openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error) {
	stream, err := getClientByModel(ctx, req.Model).CreateChatCompletionStream(ctx, *req)
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
				_ = usage.SaveFromText(ctx, req.Model, sb.String())
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

func (s *OpenAI) CreateCompletion(ctx context.Context, req *openai.CompletionRequest) (openai.CompletionResponse, error) {
	resp, err := getClientByModel(ctx, req.Model).CreateCompletion(ctx, *req)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err.Error())
		return openai.CompletionResponse{}, err
	}
	_ = usage.Save(ctx, req.Model, resp.Usage.TotalTokens)
	return resp, nil
}

func (s *OpenAI) CreateCompletionStream(ctx context.Context, req *openai.CompletionRequest, dataChan chan openai.CompletionResponse, errChan chan error) {
	stream, err := getClientByModel(ctx, req.Model).CreateCompletionStream(ctx, *req)
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
				_ = usage.SaveFromText(ctx, req.Model, sb.String())
				errChan <- err
				return
			}
			slog.ErrorContext(ctx, "chat stream receive error", "err", err.Error())
			errChan <- err
			return
		}
		if len(resp.Choices) > 0 {
			sb.WriteString(resp.Choices[0].Text)
			dataChan <- resp
		}
	}
}

func (s *OpenAI) GetModels(ctx context.Context) (openai.ModelsList, error) {
	return getClientByModel(ctx, openai.GPT3Dot5Turbo).ListModels(ctx)
}

func (s *OpenAI) CreateEmbeddings(ctx context.Context, req openai.EmbeddingRequest) (openai.EmbeddingResponse, error) {
	resp, err := getClientByModel(ctx, req.Model.String()).CreateEmbeddings(ctx, req)
	return resp, err
}
