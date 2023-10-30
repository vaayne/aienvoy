package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm"
	"github.com/sashabaranov/go-openai"
)

func main() {
	//model := llmclaude.ModelClaudeV1Dot3
	//model := llmclaudeweb.ModelClaudeWeb
	model := openai.GPT3Dot5Turbo
	//model := llmbard.ModelBard
	svc := llm.New(model)

	ctx := context.Background()
	req := &openai.CompletionRequest{
		Model:  model,
		Prompt: "hello, please introduce yourself",
		Stream: true,
	}

	if req.Stream {
		dataChan := make(chan openai.CompletionResponse)
		errChan := make(chan error)
		sb := &strings.Builder{}
		go svc.CreateCompletionStream(ctx, req, dataChan, errChan)
		for {
			select {
			case data := <-dataChan:
				fmt.Print(data.Choices[0].Text)
				sb.WriteString(data.Choices[0].Text)
			case err := <-errChan:
				if errors.Is(err, io.EOF) {
					return
				}
				slog.Error("get response from llm error", "err", err)
				return
			}
		}
	} else {
		resp, err := svc.CreateCompletion(ctx, req)
		slog.Info("get response from llm success", "model", model, "resp", resp.Choices[0].Text, "err", err)
	}
}
