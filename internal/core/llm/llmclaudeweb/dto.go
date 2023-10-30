package llmclaudeweb

import (
	"fmt"
	"strings"
	"time"

	"github.com/Vaayne/aienvoy/pkg/claudeweb"
	"github.com/sashabaranov/go-openai"
)

func buildPrompt(req *openai.ChatCompletionRequest) string {
	sb := strings.Builder{}
	for _, message := range req.Messages {
		sb.WriteString(fmt.Sprintf("\n\n%s: ", message.Role))
		sb.WriteString(message.Content)
	}

	return sb.String()
}

type ClaudeResponse struct {
	claudeweb.ChatMessageResponse
}

func (cr *ClaudeResponse) stopReasonMaapping() string {
	switch cr.StopReason {
	case "stop_sequence":
		return "stop"
	case "max_tokens":
		return "length"
	default:
		return cr.StopReason
	}
}

func (cr *ClaudeResponse) ToOpenAIChatCompletionResponse() openai.ChatCompletionResponse {
	return openai.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", cr.LogId),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: cr.Completion,
				},
				FinishReason: openai.FinishReason(cr.stopReasonMaapping()),
			},
		},
	}
}

func (cr *ClaudeResponse) ToOpenAIChatCompletionStreamResponse() openai.ChatCompletionStreamResponse {
	return openai.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", cr.LogId),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []openai.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Role:    "assistant",
					Content: cr.Completion,
				},
				FinishReason: openai.FinishReason(cr.stopReasonMaapping()),
			},
		},
	}
}

func (cr *ClaudeResponse) ToOpenAICompletionResponse() openai.CompletionResponse {
	return openai.CompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", cr.LogId),
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Choices: []openai.CompletionChoice{
			{
				Index:        0,
				Text:         cr.Completion,
				FinishReason: cr.stopReasonMaapping(),
			},
		},
	}
}
