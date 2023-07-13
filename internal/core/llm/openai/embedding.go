package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type EmbeddingRequest struct {
	openai.EmbeddingRequest
}

type EmbeddingResponse struct {
	openai.EmbeddingResponse
}

func (s *OpenAI) CreateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	resp, err := getClientByModel(req.Model.String()).CreateEmbeddings(ctx, req)
	return &EmbeddingResponse{resp}, err
}
