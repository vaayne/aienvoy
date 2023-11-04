package llm

import (
	"context"
	"log/slog"

	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/Vaayne/aienvoy/pkg/llm/bard"
	"github.com/Vaayne/aienvoy/pkg/llm/claude"
	"github.com/Vaayne/aienvoy/pkg/llm/claudeweb"
	"github.com/Vaayne/aienvoy/pkg/llm/openai"
	"github.com/Vaayne/aienvoy/pkg/llm/phind"
	"github.com/pocketbase/pocketbase/daos"
)

type Service interface {
	ListModels() []string
	CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error)
	CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, respChan chan llm.ChatCompletionStreamResponse, errChan chan error)

	CreateConversation(ctx context.Context, name string) (llm.Conversation, error)
	ListConversations(ctx context.Context) ([]llm.Conversation, error)
	GetConversation(ctx context.Context, id string) (llm.Conversation, error)
	DeleteConversation(ctx context.Context, id string) error

	CreateMessageStream(ctx context.Context, conversationId string, req llm.ChatCompletionRequest, respChan chan llm.ChatCompletionStreamResponse, errChan chan error)
	CreateMessage(ctx context.Context, conversationId string, req llm.ChatCompletionRequest) (llm.Message, error)
	ListMessages(ctx context.Context, conversationId string) ([]llm.Message, error)
	GetMessage(ctx context.Context, id string) (llm.Message, error)
	DeleteMessage(ctx context.Context, id string) error
}

var modelClientMappings map[string]func() Service

func init() {
	modelClientMappings = make(map[string]func() Service)

	// bard
	createBard := func() Service {
		return newBard()
	}
	for _, model := range bard.ListModels() {
		modelClientMappings[model] = createBard
	}

	// claude
	createClaude := func() Service {
		return newClaude()
	}
	for _, model := range claude.ListModels() {
		modelClientMappings[model] = createClaude
	}

	// claude web
	createClaudeWeb := func() Service {
		return newClaudeWeb()
	}
	for _, model := range claudeweb.ListModels() {
		modelClientMappings[model] = createClaudeWeb
	}

	// openai
	createOpenai := func() Service {
		return newOpenai()
	}
	for _, model := range openai.ListModels() {
		modelClientMappings[model] = createOpenai
	}

	// phind
	createPhind := func() Service {
		return newPhind()
	}
	for _, model := range phind.ListModels() {
		modelClientMappings[model] = createPhind
	}
}

func newDao() llm.Dao {
	return llm.DefaultDao
}

func New(model string) Service {
	createCli, ok := modelClientMappings[model]
	if ok {
		return createCli()
	}
	slog.Error("unknown model", "model", model)
	return nil
}
