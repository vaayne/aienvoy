package llmopenai

import (
	"context"
	"log/slog"

	"github.com/Vaayne/aienvoy/internal/core/llm"
	"github.com/Vaayne/aienvoy/internal/pkg/ctxutils"
	"github.com/pocketbase/pocketbase/daos"

	"github.com/pkoukk/tiktoken-go"
	openai "github.com/sashabaranov/go-openai"
)

type TikToken struct {
	*tiktoken.Tiktoken
}

func NewTikToken(model string) *TikToken {
	tk, err := tiktoken.EncodingForModel(model)
	if err != nil {
		slog.Error("tiktoken.EncodingForModel", "err", err)
	}

	return &TikToken{
		Tiktoken: tk,
	}
}

func (t *TikToken) Encode(text string) int {
	return len(t.Tiktoken.Encode(text, nil, nil))
}

func (t *TikToken) CalculateTotalTokensFromMessages(messages []openai.ChatCompletionMessage) int {
	totalTokens := 0
	for _, message := range messages {
		totalTokens += t.Encode(message.Content)
	}
	return totalTokens
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
	slog.InfoContext(ctx, "save llm token usage", "token", tokenUsage, "model", model)
	return nil
}
