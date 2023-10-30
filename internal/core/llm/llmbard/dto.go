package llmbard

import (
	"fmt"
	"strings"
	"time"

	"github.com/Vaayne/aienvoy/pkg/bard"
	"github.com/sashabaranov/go-openai"
)

type BardAnswer struct {
	bard.Answer
}

func (ba *BardAnswer) ToOpenAIChatCompletionResponse() openai.ChatCompletionResponse {
	return openai.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s:%s:%s", ba.ConversationID, ba.ResponseID, ba.Choices[0].ID),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: ba.Choices[0].Content,
				},
				FinishReason: "stop",
			},
		},
	}
}

func (ba *BardAnswer) ToOpenAIChatCompletionStreamResponse() openai.ChatCompletionStreamResponse {
	return openai.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s:%s:%s", ba.ConversationID, ba.ResponseID, ba.Choices[0].ID),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []openai.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Role:    "assistant",
					Content: ba.Choices[0].Content,
				},
				FinishReason: "stop",
			},
		},
	}
}

func (ba *BardAnswer) ToOpenAICompletionResponse() openai.CompletionResponse {
	return openai.CompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s:%s:%s", ba.ConversationID, ba.ResponseID, ba.Choices[0].ID),
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Choices: []openai.CompletionChoice{
			{
				Index:        0,
				Text:         ba.Choices[0].Content,
				FinishReason: "stop",
			},
		},
	}
}

func buildPrompt(req *openai.ChatCompletionRequest) string {
	sb := strings.Builder{}
	for _, message := range req.Messages {
		sb.WriteString(fmt.Sprintf("\n\n%s: ", message.Role))
		sb.WriteString(message.Content)
	}

	return sb.String()
}
