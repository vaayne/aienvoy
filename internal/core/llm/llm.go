package llm

import (
	"context"
	"log/slog"

	"github.com/Vaayne/aienvoy/internal/core/llm/llmbard"
	"github.com/Vaayne/aienvoy/internal/core/llm/llmclaude"
	"github.com/Vaayne/aienvoy/internal/core/llm/llmclaudeweb"
	"github.com/Vaayne/aienvoy/internal/core/llm/llmopenai"

	"github.com/sashabaranov/go-openai"
)

type Service interface {
	CreateChatCompletion(ctx context.Context, req *openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
	CreateChatCompletionStream(ctx context.Context, req *openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error)

	CreateCompletion(ctx context.Context, req *openai.CompletionRequest) (openai.CompletionResponse, error)
	CreateCompletionStream(ctx context.Context, req *openai.CompletionRequest, dataChan chan openai.CompletionResponse, errChan chan error)
}

func New(model string) Service {
	switch model {
	// openai base model
	case openai.GPT3Dot5Turbo, openai.GPT3Dot5Turbo16K, openai.GPT4, openai.GPT432K, openai.GPT3Dot5TurboInstruct:
		return llmopenai.New()
	// openai time limited model
	case openai.GPT3Dot5Turbo0301, openai.GPT3Dot5Turbo0613, openai.GPT3Dot5Turbo16K0613, openai.GPT40314, openai.GPT40613, openai.GPT432K0314, openai.GPT432K0613:
		return llmopenai.New()
	// claude models
	case llmclaude.ModelClaudeV2, llmclaude.ModelClaudeV1Dot3, llmclaude.ModelClaudeInstantV1Dot2:
		return llmclaude.New()
	case llmclaudeweb.ModelClaudeWeb:
		return llmclaudeweb.New()
	case llmbard.ModelBard:
		return llmbard.New()
	default:
		slog.Error("unknown model", "model", model)
		return nil
	}
}
