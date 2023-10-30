package llmopenai

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm/dto"
	"github.com/Vaayne/aienvoy/internal/core/llm/usage"

	"github.com/sashabaranov/go-openai"
)

// OpenAI is a client for AI services
type OpenAI struct{}

func New() *OpenAI {
	return &OpenAI{}
}

func (s *OpenAI) CreateChatCompletion(ctx context.Context, req *openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	slog.InfoContext(ctx, "chat with OpenAI start")
	resp, err := getClientByModel(ctx, req.Model).CreateChatCompletion(ctx, *req)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err)
		return openai.ChatCompletionResponse{}, err
	}
	_ = usage.Save(ctx, req.Model, resp.Usage.TotalTokens)
	slog.InfoContext(ctx, "chat with OpenAI success")
	return resp, nil
}

func (s *OpenAI) CreateChatCompletionStream(ctx context.Context, req *openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat with OpenAI stream start")
	stream, err := getClientByModel(ctx, req.Model).CreateChatCompletionStream(ctx, *req)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err)
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
				slog.InfoContext(ctx, "chat with OpenAI stream success")
				errChan <- err
				return
			}
			slog.ErrorContext(ctx, "chat with OpenAI stream error", "err", err)
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
	chatReq := dto.CompletionRequestToChatCompletionRequest(*req)
	resp, err := s.CreateChatCompletion(ctx, &chatReq)
	if err != nil {
		return openai.CompletionResponse{}, err
	}
	return dto.ChatCompletionResponseToCompletionResponse(resp), nil
}

func (s *OpenAI) CreateCompletionStream(ctx context.Context, req *openai.CompletionRequest, dataChan chan openai.CompletionResponse, errChan chan error) {
	chatReq := dto.CompletionRequestToChatCompletionRequest(*req)
	respChan := make(chan openai.ChatCompletionStreamResponse)
	innerErrorChan := make(chan error)

	go s.CreateChatCompletionStream(ctx, &chatReq, respChan, innerErrorChan)

	for {
		select {
		case resp := <-respChan:
			dataChan <- dto.ChatCompletionStreamResponseToCompletionResponse(resp)
		case err := <-innerErrorChan:
			errChan <- err
			return
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
