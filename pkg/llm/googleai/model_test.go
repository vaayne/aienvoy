package googleai

import (
	"reflect"
	"testing"

	"github.com/Vaayne/aienvoy/pkg/llm"
)

func TestFromChatCompletionRequest(t *testing.T) {
	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatCompletionMessage{
			{Role: llm.ChatMessageRoleSystem, Content: "System message"},
			{Role: llm.ChatMessageRoleUser, Content: "How are you?"},
			{Role: llm.ChatMessageRoleAssistant, Content: "I'm fine, thank you."},
			{Role: llm.ChatMessageRoleUser, Content: "Hello"},
		},
		N:           5,
		Stop:        []string{"stop"},
		Temperature: 0.5,
		TopP:        0.9,
		MaxTokens:   100,
	}

	expected := ChatRequest{
		Contents: []ChatMessage{
			{Role: "user", Parts: []ChatMessagePart{{Text: "System message"}}},
			{Role: "model", Parts: []ChatMessagePart{{Text: ""}}},
			{Role: "user", Parts: []ChatMessagePart{{Text: "How are you?"}}},
			{Role: "model", Parts: []ChatMessagePart{{Text: "I'm fine, thank you."}}},
			{Role: "user", Parts: []ChatMessagePart{{Text: "Hello"}}},
		},
		SafetySettings: []SafetySetting{
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: "BLOCK_ONLY_HIGH",
			},
		},
		GenerationConfig: GenerationConfig{
			Stop:        req.Stop,
			Temperature: req.Temperature,
			TopP:        req.TopP,
			TopK:        5,
			MaxTokens:   req.MaxTokens,
		},
	}

	r := ChatRequest{}
	result := r.FromChatCompletionRequest(req)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FromChatCompletionRequest() = %v, want %v", result, expected)
	}
}
