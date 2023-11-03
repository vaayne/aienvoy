package openai

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/sashabaranov/go-openai"
)

const (
	ModelGPT432K          = "gpt-4-32k"
	ModelGPT4             = "gpt-4"
	ModelGPT3Dot5Turbo16K = "gpt-3.5-turbo-16k"
	ModelGPT3Dot5Turbo    = "gpt-3.5-turbo"
)

type OpenAI struct {
	*llm.LLM
}

func New(cfg openai.ClientConfig, dao llm.Dao) *OpenAI {
	return &OpenAI{
		llm.New(dao, NewClient(cfg)),
	}
}

func ListModels() []string {
	return []string{
		ModelGPT432K, ModelGPT4, ModelGPT3Dot5Turbo16K, ModelGPT3Dot5Turbo,
	}
}

func (s *Client) ListModels() []string {
	return ListModels()
}

func (s *Client) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	slog.InfoContext(ctx, "chat with OpenAI start")
	openaiReq := toOpenAIChatCompletionRequest(req)
	resp, err := s.Client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err)
		return llm.ChatCompletionResponse{}, err
	}
	slog.InfoContext(ctx, "chat with OpenAI success")
	return toLLMChatCompletionResponse(resp), nil
}

func (s *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat with OpenAI stream start")
	openaiReq := toOpenAIChatCompletionRequest(req)
	stream, err := s.Client.CreateChatCompletionStream(ctx, openaiReq)
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
			dataChan <- toLLMChatCompletionStreamResponse(resp)
		}
	}
}
