package llm

import "context"

type Interface interface {
	ListModels() []string
	CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error)
	CreateChatCompletionStream(ctx context.Context, req ChatCompletionRequest, respChan chan ChatCompletionStreamResponse, errChan chan error)

	CreateConversation(ctx context.Context, name string) (Conversation, error)
	ListConversations(ctx context.Context) ([]Conversation, error)
	GetConversation(ctx context.Context, id string) (Conversation, error)
	DeleteConversation(ctx context.Context, id string) error

	CreateMessageStream(ctx context.Context, conversationId string, req ChatCompletionRequest, respChan chan ChatCompletionStreamResponse, errChan chan error)
	CreateMessage(ctx context.Context, conversationId string, req ChatCompletionRequest) (Message, error)
	ListMessages(ctx context.Context, conversationId string) ([]Message, error)
	GetMessage(ctx context.Context, id string) (Message, error)
	DeleteMessage(ctx context.Context, id string) error
}
