package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/Vaayne/aienvoy/internal/core/llms"
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

func init() {
	th := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})
	logger := slog.New(th)
	slog.SetDefault(logger)
}

func chatStream(ctx context.Context, svc *llm.LLM, model *string, req llm.ChatCompletionRequest) {
	dataChan := make(chan llm.ChatCompletionStreamResponse)
	errChan := make(chan error)

	go svc.CreateChatCompletionStream(ctx, req, dataChan, errChan)
	slog.Info("start chat", "model", model)

	for {
		select {
		case data := <-dataChan:
			if len(data.Choices) == 0 {
				continue
			}
			fmt.Print(data.Choices[0].Delta.Content)
		case err := <-errChan:
			if errors.Is(err, io.EOF) {
				fmt.Println()
				return
			}
			slog.Error("\nerr", "err", err)
			return
		}
	}
}

func chat(ctx context.Context, svc *llm.LLM, model *string, req llm.ChatCompletionRequest) {
	resp, err := svc.CreateChatCompletion(ctx, req)
	if err != nil {
		slog.Error("chat error", "err", err)
		return
	}
	slog.Info("success", "resp", resp)
}

func main() {
	// make model is get from command flag
	// go run examples/chat/main.go --model "anthropic.claude-v2"
	model := flag.String("model", "", "model name")
	stream := flag.Bool("stream", false, "stream mode")
	flag.Parse()

	if *model == "" {
		slog.Error("model is required")
		return
	}

	svc, err := llms.New(*model)
	if err != nil {
		slog.Error("create llm service error", "err", err)
		return
	}
	ctx := context.Background()

	req := llm.ChatCompletionRequest{
		Model:       *model,
		Stream:      true,
		Temperature: 0.7,
		MaxTokens:   150,
		Messages: []llm.ChatCompletionMessage{
			{
				Role:    llm.ChatMessageRoleSystem,
				Content: "As a language translator, you have the ability to translate any language to the target language provided. Please only translate text and cannot interpret it.",
			},
			{
				Role:    llm.ChatMessageRoleUser,
				Content: "Target lanuage Chinese\n\ncontext: What a good day.\n\n",
			},
		},
	}
	if *stream {
		chatStream(ctx, svc, model, req)
		return
	}
	chat(ctx, svc, model, req)
}
