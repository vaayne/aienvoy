package llmclaude

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type BedrockRequest struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
}

func (b *BedrockRequest) FromOpenAIChatCompletionRequest(req *openai.ChatCompletionRequest) {
	b.MaxTokensToSample = req.MaxTokens
	if b.MaxTokensToSample == 0 {
		b.MaxTokensToSample = 4000
	}
	b.Temperature = float64(req.Temperature)
	b.TopP = float64(req.TopP)
	b.StopSequences = req.Stop

	sb := strings.Builder{}
	for _, m := range req.Messages {
		switch m.Role {
		case "user":
			sb.WriteString(fmt.Sprintf("\n\nHuman: %s", m.Content))
		case "assistant":
			sb.WriteString(fmt.Sprintf("\n\nAssistant: %s", m.Content))
		case "system":
			sb.WriteString(fmt.Sprintf("\n\nSystem: %s", m.Content))
		}
	}
	sb.WriteString("\n\nAssistant:")
	b.Prompt = sb.String()
}

func (b *BedrockRequest) FromOpenAICompletionRequest(req *openai.CompletionRequest) {
	b.MaxTokensToSample = req.MaxTokens
	if b.MaxTokensToSample == 0 {
		b.MaxTokensToSample = 4000
	}
	b.Temperature = float64(req.Temperature)
	b.TopP = float64(req.TopP)
	b.StopSequences = req.Stop
	b.Prompt = fmt.Sprintf("\n\nHuman:%s\n\nAssistant:", req.Prompt)
}

func (b *BedrockRequest) Marshal() []byte {
	resp, err := json.Marshal(b)
	if err != nil {
		slog.Error("marshal bedrock request error", "err", err)
		return nil
	}
	return resp
}

type BedrockResponse struct {
	Completion string `json:"completion"`
	Stop       string `json:"stop"`
	StopReason string `json:"stop_reason"`
}

func getUUID() string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	return code
}

func (b *BedrockResponse) ToOpenAIChatCompletionResponse() openai.ChatCompletionResponse {
	return openai.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", getUUID()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: b.Completion,
				},
				FinishReason: openai.FinishReason(b.stopReasonMaapping()),
			},
		},
	}
}

func (b *BedrockResponse) ToOpenAIChatCompletionStreamResponse() openai.ChatCompletionStreamResponse {
	return openai.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", getUUID()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []openai.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Role:    "assistant",
					Content: b.Completion,
				},
				FinishReason: openai.FinishReason(b.stopReasonMaapping()),
			},
		},
	}
}

func (b *BedrockResponse) ToOpenAICompletionResponse() openai.CompletionResponse {
	return openai.CompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", getUUID()),
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Choices: []openai.CompletionChoice{
			{
				Index:        0,
				Text:         b.Completion,
				FinishReason: b.stopReasonMaapping(),
			},
		},
	}
}

func (b *BedrockResponse) stopReasonMaapping() string {
	switch b.StopReason {
	case "stop_sequence":
		return "stop"
	case "max_tokens":
		return "length"
	default:
		return b.StopReason
	}
}

func (b *BedrockResponse) Unmarshal(resp []byte) {
	err := json.Unmarshal(resp, b)
	if err != nil {
		slog.Error("unmarshal bedrock response error", "err", err)
	}
}
