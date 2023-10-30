package usage

import (
	"context"
	"log/slog"

	"github.com/sashabaranov/go-openai"

	"github.com/Vaayne/aienvoy/internal/pkg/ctxutils"
	"github.com/pocketbase/pocketbase/daos"

	"github.com/pkoukk/tiktoken-go"
)

type Token struct {
	*tiktoken.Tiktoken
}

func NewTikToken(model string) *Token {
	tk, err := tiktoken.EncodingForModel(model)
	if err != nil {
		slog.Warn("init tiktoken error, fallback to default model", "err", err)
		// if the model not supported, default to base model
		tk, _ = tiktoken.EncodingForModel(openai.GPT3Dot5Turbo)
	}

	return &Token{
		Tiktoken: tk,
	}
}

func (t *Token) Encode(text string) int {
	return len(t.Tiktoken.Encode(text, nil, nil))
}

func (t *Token) CalculateTotalTokensFromMessages(messages []openai.ChatCompletionMessage) int {
	totalTokens := 0
	for _, message := range messages {
		totalTokens += t.Encode(message.Content)
	}
	return totalTokens
}

func SaveFromMessages(ctx context.Context, model string, messages []openai.ChatCompletionMessage) error {
	tk := NewTikToken(model)
	tokenUsage := tk.CalculateTotalTokensFromMessages(messages)
	return Save(ctx, model, tokenUsage)
}

func SaveFromText(ctx context.Context, model string, text string) error {
	tk := NewTikToken(model)
	tokenUsage := tk.Encode(text)
	return Save(ctx, model, tokenUsage)
}

func Save(ctx context.Context, model string, usage int) error {
	usageDao := ctxutils.GetDao(ctx)
	if usageDao == nil {
		slog.Error("save llm token usage error, dao is nil")
		return nil
	}

	if err := usageDao.RunInTransaction(func(tx *daos.Dao) error {
		return SaveLlmUsage(ctx, tx, &LlmUsages{
			UserId:     ctxutils.GetUserId(ctx),
			ApiKey:     ctxutils.GetApiKey(ctx),
			TokenUsage: int64(usage),
			Model:      model,
		})
	}); err != nil {
		slog.ErrorContext(ctx, "save llm token usage error", "err", err.Error())
		return err
	}
	slog.InfoContext(ctx, "save llm token usage", "token", usage, "model", model)
	return nil
}
