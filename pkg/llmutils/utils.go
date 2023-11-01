package llmutils

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

func CreateCompletion(ctx context.Context, req *openai.CompletionRequest,
	createChatCompletion func(context.Context, *openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error),
) (openai.CompletionResponse, error) {
	chatReq := CompletionRequestToChatCompletionRequest(*req)
	resp, err := createChatCompletion(ctx, &chatReq)
	if err != nil {
		return openai.CompletionResponse{}, err
	}
	return ChatCompletionResponseToCompletionResponse(resp), nil
}

func CreateCompletionStream(ctx context.Context, req *openai.CompletionRequest, dataChan chan openai.CompletionResponse, errChan chan error,
	createChatCompletionStream func(context.Context, *openai.ChatCompletionRequest, chan openai.ChatCompletionStreamResponse, chan error),
) {
	chatReq := CompletionRequestToChatCompletionRequest(*req)
	respChan := make(chan openai.ChatCompletionStreamResponse)
	innerErrorChan := make(chan error)

	go createChatCompletionStream(ctx, &chatReq, respChan, innerErrorChan)

	for {
		select {
		case resp := <-respChan:
			dataChan <- ChatCompletionStreamResponseToCompletionResponse(resp)
		case err := <-innerErrorChan:
			errChan <- err
			return
		}
	}
}
