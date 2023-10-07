package llmopenai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

func (s *OpenAI) CreateEmbeddings(ctx context.Context, req openai.EmbeddingRequest) (openai.EmbeddingResponse, error) {
	resp, err := getClientByModel(req.Model.String()).CreateEmbeddings(ctx, req)
	return resp, err
}
