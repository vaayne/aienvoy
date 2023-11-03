package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

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
	req := llm.ChatCompletionRequest{
		Model: model,
		Messages: []llm.ChatCompletionMessage{
			{
				Content: "what's the latest news",
				Role:    llm.ChatMessageRoleUser,
			},
		},
		Stream: true,
	}

	if req.Stream {
		dataChan := make(chan llm.ChatCompletionStreamResponse)
		errChan := make(chan error)
		sb := &strings.Builder{}
		go svc.CreateChatCompletionStream(ctx, req, dataChan, errChan)
		for {
			select {
			case data := <-dataChan:
				fmt.Print(data.Choices[0].Delta.Content)
				sb.WriteString(data.Choices[0].Delta.Content)
			case err := <-errChan:
				if errors.Is(err, io.EOF) {
					return
				}
				slog.Error("get response from llm error", "err", err)
				return
			}
		}
	} else {
		resp, err := svc.CreateChatCompletion(ctx, req)
		slog.Info("get response from llm success", "model", model, "resp", resp.Choices[0].Message.Content, "err", err)
	}
}
