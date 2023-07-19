package openai

import (
	"time"

	"aienvoy/internal/pkg/context"
	"aienvoy/internal/pkg/dao"
	"aienvoy/internal/pkg/logger"

	"github.com/pocketbase/pocketbase/daos"
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
	if err != nil {
		logger.GetSugaredLoggerWithContext(ctx).Errorw("chat with OpenAI error", "err", err.Error())
		return ChatCompletionResponse{}, err
	}

	// save usage
	usage := &dao.Usage{
		UserId:   ctx.UserId(),
		ApiKey:   ctx.APIKey(),
		Usage:    resp.Usage.TotalTokens,
		Model:    req.Model,
		DateTime: getCurrentTimeTruncatedToHour(),
	}
	d := dao.New(ctx)
	if err := d.RunInTransaction(
		func(tx *daos.Dao) error {
			return d.CreateUsage(tx, usage)
		}); err != nil {
		logger.GetSugaredLoggerWithContext(ctx).Errorw("save usage error", "err", err.Error())
		return ChatCompletionResponse{}, err
	}
	return ChatCompletionResponse{resp}, nil
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

func getCurrentTimeTruncatedToHour() time.Time {
	return time.Now().UTC().Truncate(time.Hour)
}
