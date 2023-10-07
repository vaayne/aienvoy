package llmopenai

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm"
	"github.com/Vaayne/aienvoy/internal/pkg/ctxutils"

	"github.com/pocketbase/pocketbase/daos"
	"github.com/sashabaranov/go-openai"
)

type ListModelsResponse struct {
	openai.ModelsList
}

func (s *OpenAI) Chat(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	resp, err := getClientByModel(req.Model).CreateChatCompletion(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err.Error())
		return openai.ChatCompletionResponse{}, err
	}

	_ = saveUsage(ctx, req.Model, resp.Usage.TotalTokens)

	return resp, nil
}

func (s *OpenAI) ChatStream(ctx context.Context, req openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error) {
	stream, err := getClientByModel(req.Model).CreateChatCompletionStream(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err.Error())
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
				_ = saveUsage(ctx, req.Model, totalTokens)
				errChan <- err
				return
			}
			slog.ErrorContext(ctx, "chat stream receive error", "err", err.Error())
			errChan <- err
			return
		}
		if len(resp.Choices) > 0 {
			sb.WriteString(resp.Choices[0].Delta.Content)
			dataChan <- resp
		}
	}
}

func (s *OpenAI) GetModels(ctx context.Context) (ListModelsResponse, error) {
	resp, err := getClientByModel(openai.GPT3Dot5Turbo).ListModels(ctx)
	return ListModelsResponse{resp}, err
}

func saveUsage(ctx context.Context, model string, tokenUsage int) error {
	usageDao := ctxutils.GetDao(ctx)

	if err := usageDao.RunInTransaction(
		func(tx *daos.Dao) error {
			return llm.SaveLlmUsage(ctx, tx, &llm.LlmUsages{
				UserId:     ctxutils.GetUserId(ctx),
				ApiKey:     ctxutils.GetApiKey(ctx),
				TokenUsage: int64(tokenUsage),
				Model:      model,
			})
		}); err != nil {
		slog.ErrorContext(ctx, "save usage error", "err", err.Error())
		return err
	}
	slog.InfoContext(ctx, "save llm token usgae", "token", tokenUsage, "model", model)
	return nil
}
