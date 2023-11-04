package main

import (
	"context"
	"log/slog"

	innerllm "github.com/Vaayne/aienvoy/internal/core/llm"
	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/Vaayne/aienvoy/pkg/llm/openai"
)

func main() {
	// bard.ModelBard
	// claude.ModelClaudeV1Dot3
	// claudeweb.ModelClaudeWeb
	// openai.ModelGPT3Dot5Turbo
	// phind.ModelPhindV1
	model := openai.ModelGPT3Dot5Turbo
	svc := innerllm.New(model)
	ctx := context.Background()

	cov, err := svc.CreateConversation(ctx, "test")
	if err != nil {
		slog.Error("create conversation error", "err", err)
		return
	}
	slog.Info("create conversation success", "conversation", cov)

	req := llm.ChatCompletionRequest{
		Model: model,
		Messages: []llm.ChatCompletionMessage{
			{
				Content: "what's the latest news",
				Role:    llm.ChatMessageRoleUser,
			},
		},
		Stream: false,
	}

	resp, err := svc.CreateChatCompletion(ctx, req)
	if err != nil {
		slog.Error("create chat completion error", "err", err)
		return
	}
	slog.Info("get response from llm success", "model", model, "resp", resp.Choices[0].Message.Content)

	covs, err := svc.ListConversations(ctx)
	if err != nil {
		slog.Error("list conversations error", "err", err)
		return
	}
	slog.Info("list conversations success", "conversations", covs)

	msgs, err := svc.ListMessages(ctx, cov.Id)
	if err != nil {
		slog.Error("list messages error", "err", err)
		return
	}
	slog.Info("list messages success", "messages", msgs)
}
