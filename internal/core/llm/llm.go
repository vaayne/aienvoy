package llm

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Service interface {
	CreateChatCompletion(ctx context.Context, req *openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
	CreateChatCompletionStream(ctx context.Context, req *openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error)

	CreateCompletion(ctx context.Context, req *openai.CompletionRequest) (openai.CompletionResponse, error)
	CreateCompletionStream(ctx context.Context, req *openai.CompletionRequest, dataChan chan openai.CompletionResponse, errChan chan error)
}
