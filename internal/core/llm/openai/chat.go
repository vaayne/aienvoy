package openai

import (
	"errors"
	"io"
	"strings"

	"aienvoy/internal/core/llm"
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

type ChatCompletionStreamResponse struct {
	openai.ChatCompletionStreamResponse
}

type ListModelsResponse struct {
	openai.ModelsList
}

func (s *OpenAI) Chat(ctx context.Context, req *ChatCompletionRequest) (ChatCompletionResponse, error) {
	resp, err := getClientByModel(req.Model).CreateChatCompletion(ctx, req.ChatCompletionRequest)
	if err != nil {
		logger.GetSugaredLoggerWithContext(ctx).Errorw("chat with OpenAI error", "err", err.Error())
		return ChatCompletionResponse{}, err
	}

	if err := saveUsage(ctx, req.Model, resp.Usage.TotalTokens); err != nil {
		logger.GetSugaredLoggerWithContext(ctx).Errorw("save usage error", "err", err.Error())
		return ChatCompletionResponse{}, err
	}

	return ChatCompletionResponse{resp}, nil
}

func (s *OpenAI) ChatStream(ctx context.Context, req *ChatCompletionRequest, dataChan chan ChatCompletionStreamResponse, errChan chan error) {
	Logger := logger.GetSugaredLoggerWithContext(ctx)

	stream, err := getClientByModel(req.Model).CreateChatCompletionStream(ctx, req.ChatCompletionRequest)
	if err != nil {
		Logger.Errorw("chat with OpenAI error", "err", err.Error())
		errChan <- err
		return
	}

	// log chat stream response
	sb := &strings.Builder{}
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				tk := NewTiktoken(req.Model)
				totalTokens := tk.Encode(sb.String())
				if err := saveUsage(ctx, req.Model, totalTokens); err != nil {
					Logger.Errorw("save usage error", "err", err.Error())
				}
				errChan <- err
				return
			}
			Logger.Errorw("chat stream receive error", "err", err.Error())
			errChan <- err
			return
		}
		sb.WriteString(resp.Choices[0].Delta.Content)
		dataChan <- ChatCompletionStreamResponse{resp}
	}
}

func (s *OpenAI) GetModels(ctx context.Context) (ListModelsResponse, error) {
	resp, err := getClientByModel(openai.GPT3Dot5Turbo).ListModels(ctx)
	return ListModelsResponse{resp}, err
}

func saveUsage(ctx context.Context, model string, tokenUsage int) error {
	usageDao := dao.New(ctx)

	if err := usageDao.RunInTransaction(
		func(tx *daos.Dao) error {
			return llm.SaveLlmUsage(ctx, tx, &llm.LlmUsages{
				UserId:     ctx.UserId(),
				ApiKey:     ctx.APIKey(),
				TokenUsage: int64(tokenUsage),
				Model:      model,
			})
		}); err != nil {
		logger.GetSugaredLoggerWithContext(ctx).Errorw("save usage error", "err", err.Error())
		return err
	}
	return nil
}
