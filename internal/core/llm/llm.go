package llm

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Service interface {
	// CreateCompletion one time chat with LLM
	CreateCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
	// CreateCompletionStream one time chat with LLM and stream response
	CreateCompletionStream(ctx context.Context, req openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error)
	// Chat with conversation history
	Chat(ctx context.Context, conversationId string, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
	// ChatStream with conversation history and stream response
	ChatStream(ctx context.Context, conversationId string, req openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error)
	// GetModels get all models
	GetModels(ctx context.Context) (openai.ModelsList, error)
	// CreateEmbeddings embed content
	CreateEmbeddings(ctx context.Context, req openai.EmbeddingRequest) (openai.EmbeddingResponse, error)
}
