package openai

import (
	"log/slog"

	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

type Tiktoken struct {
	*tiktoken.Tiktoken
}

func NewTiktoken(model string) *Tiktoken {
	tk, err := tiktoken.EncodingForModel(model)
	if err != nil {
		slog.Error("tiktoken.EncodingForModel", "err", err)
	}

	return &Tiktoken{
		Tiktoken: tk,
	}
}

func (t *Tiktoken) Encode(text string) int {
	return len(t.Tiktoken.Encode(text, nil, nil))
}

func (t *Tiktoken) CalculateTotalTokensFromMessages(messages []openai.ChatCompletionMessage) int {
	totalTokens := 0
	for _, message := range messages {
		totalTokens += t.Encode(message.Content)
	}
	return totalTokens
}
