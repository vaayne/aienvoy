package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type ChatCompletionRequest struct {
	openai.ChatCompletionRequest
}

type ChatCompletionResponse struct {
	openai.ChatCompletionResponse
}

type ChatCompletionStream struct {
	*openai.ChatCompletionStream
}

func (s *OpenAI) Chat(ctx context.Context, req *ChatCompletionRequest) (ChatCompletionResponse, error) {
	resp, err := getClientByModel(req.Model).CreateChatCompletion(ctx, req.ChatCompletionRequest)
	return ChatCompletionResponse{resp}, err
}

func (s *OpenAI) ChatStream(ctx context.Context, req *ChatCompletionRequest) (ChatCompletionStream, error) {
	resp, err := getClientByModel(req.Model).CreateChatCompletionStream(ctx, req.ChatCompletionRequest)
	return ChatCompletionStream{resp}, err
}

type ListModelsResponse struct {
	openai.ModelsList
}

func (s *OpenAI) GetModels(ctx context.Context) (ListModelsResponse, error) {
	resp, err := getClientByModel(openai.GPT3Dot5Turbo).ListModels(ctx)
	return ListModelsResponse{resp}, err
}
