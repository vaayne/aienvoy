package llm

import "context"

type Dao interface {
	SaveConversation(ctx context.Context, conversation Conversation) (Conversation, error)
	GetConversation(ctx context.Context, id string) (Conversation, error)
	ListConversations(ctx context.Context) ([]Conversation, error)
	DeleteConversation(ctx context.Context, id string) error

	SaveMessage(ctx context.Context, message Message) (Message, error)
	GetMessage(ctx context.Context, id string) (Message, error)
	ListMessages(ctx context.Context, conversationId string) ([]Message, error)
	DeleteMessage(ctx context.Context, id string) error
}
