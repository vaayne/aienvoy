package googleai

import (
	"fmt"
	"time"

	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/google/uuid"
)

type ChatMessagePart struct {
	Text string `json:"text"`
}

type ChatMessage struct {
	Role  string            `json:"role"`
	Parts []ChatMessagePart `json:"parts"`
}

type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GenerationConfig struct {
	Stop        []string `json:"stopSequences"`
	Temperature float32  `json:"temperature"`
	MaxTokens   int      `json:"maxOutputTokens"`
	TopP        float32  `json:"topP"`
	TopK        int      `json:"topK"`
}

type ChatRequest struct {
	Contents         []ChatMessage    `json:"contents"`
	SafetySettings   []SafetySetting  `json:"safetySettings"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
}

func (r ChatRequest) FromChatCompletionRequest(req llm.ChatCompletionRequest) ChatRequest {
	contents := make([]ChatMessage, 0, len(req.Messages))
	lastRole := ""
	for _, message := range req.Messages {
		role := message.Role
		if role == llm.ChatMessageRoleAssistant {
			role = "model"
		} else if role == llm.ChatMessageRoleSystem {
			role = "user"
		}

		if lastRole == "user" && role == "user" {
			contents = append(contents, ChatMessage{
				Role:  "model",
				Parts: []ChatMessagePart{{Text: ""}},
			})
		}

		contents = append(contents, ChatMessage{
			Role:  role,
			Parts: []ChatMessagePart{{Text: message.Content}},
		})
		lastRole = role
	}

	safetySettings := []SafetySetting{
		{
			Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
			Threshold: "BLOCK_ONLY_HIGH",
		},
	}

	topK := req.N
	if topK <= 0 {
		topK = 10
	}

	generationConfig := GenerationConfig{
		Stop:        req.Stop,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		TopK:        topK,
		MaxTokens:   req.MaxTokens,
	}

	return ChatRequest{
		Contents:         contents,
		SafetySettings:   safetySettings,
		GenerationConfig: generationConfig,
	}
}

type PromptFeedback struct {
	SafetySettings []SafetySetting `json:"safetySettings"`
}

type ChatResponseCandidate struct {
	Index          int             `json:"index"`
	Content        ChatMessage     `json:"content"`
	FinishReason   string          `json:"finishReason"`
	SafetySettings []SafetySetting `json:"safetySettings"`
}

type ChatResponse struct {
	Candidates     []ChatResponseCandidate `json:"candidates"`
	PromptFeedback PromptFeedback          `json:"promptFeedback"`
}

func (r ChatResponse) ToChatCompletionResponse() llm.ChatCompletionResponse {
	choices := make([]llm.ChatCompletionChoice, 0, len(r.Candidates))
	for _, candidate := range r.Candidates {
		choices = append(choices, llm.ChatCompletionChoice{
			Index: candidate.Index,
			Message: llm.ChatCompletionMessage{
				Role:    llm.ChatMessageRoleAssistant,
				Content: candidate.Content.Parts[0].Text,
			},
			FinishReason: llm.FinishReason(candidate.FinishReason),
		})
	}

	return llm.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", uuid.New().String()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: choices,
	}
}

func (r ChatResponse) ToChatCompletionStreamResponse() llm.ChatCompletionStreamResponse {
	choices := make([]llm.ChatCompletionStreamChoice, 0, len(r.Candidates))
	for _, candidate := range r.Candidates {
		choices = append(choices, llm.ChatCompletionStreamChoice{
			Index: candidate.Index,
			Delta: llm.ChatCompletionStreamChoiceDelta{
				Role:    llm.ChatMessageRoleAssistant,
				Content: candidate.Content.Parts[0].Text,
			},
			FinishReason: llm.FinishReason(candidate.FinishReason),
		})
	}

	return llm.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", uuid.New().String()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: choices,
	}
}
