package utils

import (
	"context"

	"github.com/Vaayne/aienvoy/internal/core/llm/dto"
	"github.com/sashabaranov/go-openai"
)

func CreateCompletion(ctx context.Context, req *openai.CompletionRequest,
	createChatCompletion func(context.Context, *openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error),
) (openai.CompletionResponse, error) {
	chatReq := dto.CompletionRequestToChatCompletionRequest(*req)
	resp, err := createChatCompletion(ctx, &chatReq)
	if err != nil {
		return openai.CompletionResponse{}, err
	}
	return dto.ChatCompletionResponseToCompletionResponse(resp), nil
}

func CreateCompletionStream(ctx context.Context, req *openai.CompletionRequest, dataChan chan openai.CompletionResponse, errChan chan error,
	createChatCompletionStream func(context.Context, *openai.ChatCompletionRequest, chan openai.ChatCompletionStreamResponse, chan error),
) {
	chatReq := dto.CompletionRequestToChatCompletionRequest(*req)
	respChan := make(chan openai.ChatCompletionStreamResponse)
	innerErrorChan := make(chan error)

	go createChatCompletionStream(ctx, &chatReq, respChan, innerErrorChan)

	for {
		select {
		case resp := <-respChan:
			dataChan <- dto.ChatCompletionStreamResponseToCompletionResponse(resp)
		case err := <-innerErrorChan:
			errChan <- err
			return
		}
	}
}
