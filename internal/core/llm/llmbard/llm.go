package llmbard

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/Vaayne/aienvoy/internal/core/llm/dto"
	"github.com/Vaayne/aienvoy/internal/core/llm/usage"
	"github.com/sashabaranov/go-openai"
)

const ModelBard = "bard"

func (b *Bard) CreateChatCompletion(ctx context.Context, req *openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	slog.InfoContext(ctx, "chat with Google Bard start")
	prompt := buildPrompt(req)
	resp, err := b.Ask(prompt, "", "", "", 0)
	if err != nil {
		slog.ErrorContext(ctx, "chat with Google Bard error", "err", err)
		return openai.ChatCompletionResponse{}, fmt.Errorf("bard got an error, %w", err)
	}
	ba := &BardAnswer{Answer: *resp}
	res := ba.ToOpenAIChatCompletionResponse()
	_ = usage.SaveFromText(ctx, req.Model, res.Choices[0].Message.Content)
	slog.InfoContext(ctx, "chat with Google Bard success")
	return res, nil
}

func (b *Bard) CreateChatCompletionStream(ctx context.Context, req *openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat with Google Bard stream start")
	prompt := buildPrompt(req)
	resp, err := b.Ask(prompt, "", "", "", 0)
	if err != nil {
		errChan <- fmt.Errorf("bard got an error, %w", err)
		slog.ErrorContext(ctx, "chat with Google Bard stream error", "err", err)
		return
	}
	ba := &BardAnswer{Answer: *resp}
	res := ba.ToOpenAIChatCompletionStreamResponse()
	_ = usage.SaveFromText(ctx, req.Model, res.Choices[0].Delta.Content)
	slog.InfoContext(ctx, "chat with Google Bard stream success")
	dataChan <- res
	errChan <- io.EOF
}

func (b *Bard) CreateCompletion(ctx context.Context, req *openai.CompletionRequest) (openai.CompletionResponse, error) {
	chatReq := dto.CompletionRequestToChatCompletionRequest(*req)
	resp, err := b.CreateChatCompletion(ctx, &chatReq)
	if err != nil {
		return openai.CompletionResponse{}, err
	}
	return dto.ChatCompletionResponseToCompletionResponse(resp), nil
}

func (b *Bard) CreateCompletionStream(ctx context.Context, req *openai.CompletionRequest, dataChan chan openai.CompletionResponse, errChan chan error) {
	chatReq := dto.CompletionRequestToChatCompletionRequest(*req)

	respChan := make(chan openai.ChatCompletionStreamResponse)
	innerErrorChan := make(chan error)

	go b.CreateChatCompletionStream(ctx, &chatReq, respChan, innerErrorChan)

	for {
		select {
		case resp := <-respChan:
			data := dto.ChatCompletionStreamResponseToCompletionResponse(resp)
			dataChan <- data
		case err := <-innerErrorChan:
			errChan <- err
			return
		}
	}
}
