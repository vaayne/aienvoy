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

var modelMappings = map[string]string{
	openai.GPT3Dot5Turbo: openai.GPT3Dot5Turbo1106,
	openai.GPT4:          openai.GPT4TurboPreview,
}

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
		openai.GPT3Dot5Turbo, openai.GPT3Dot5Turbo0301, openai.GPT3Dot5Turbo0613, openai.GPT3Dot5Turbo1106,
		openai.GPT3Dot5Turbo16K, openai.GPT3Dot5Turbo16K0613,
		openai.GPT3Dot5TurboInstruct,
		openai.GPT4, openai.GPT40314, openai.GPT40613,
		openai.GPT4TurboPreview,
		openai.GPT4VisionPreview,
		openai.GPT432K, openai.GPT432K0314, openai.GPT432K0613,
	}
}

func (s *Client) ListModels() []string {
	return ListModels()
}

func (s *Client) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	openaiReq := toOpenAIChatCompletionRequest(req)
	slog.InfoContext(ctx, "chat start", "llm", openaiReq.Model, "is_stream", openaiReq.Stream)
	resp, err := s.Client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err)
		return llm.ChatCompletionResponse{}, err
	}
	slog.InfoContext(ctx, "chat success", "llm", openaiReq.Model, "is_stream", openaiReq.Stream)

	return toLLMChatCompletionResponse(resp), nil
}

func (s *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	openaiReq := toOpenAIChatCompletionRequest(req)
	slog.InfoContext(ctx, "chat start", "llm", openaiReq.Model, "is_stream", openaiReq.Stream)
	stream, err := s.Client.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		slog.InfoContext(ctx, "chat error", "llm", openaiReq.Model, "is_stream", openaiReq.Stream, "err", err)
		errChan <- err
		return
	}

	// log chat stream response
	sb := &strings.Builder{}
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.InfoContext(ctx, "chat success", "llm", openaiReq.Model, "is_stream", openaiReq.Stream)
				errChan <- err
				return
			}
			slog.InfoContext(ctx, "chat error", "llm", openaiReq.Model, "is_stream", openaiReq.Stream, "err", err)
			errChan <- err
			return
		}
		if len(resp.Choices) > 0 {
			sb.WriteString(resp.Choices[0].Delta.Content)
			dataChan <- toLLMChatCompletionStreamResponse(resp)
		}
	}
}
