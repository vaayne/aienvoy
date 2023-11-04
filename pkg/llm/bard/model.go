package bard

import (
	"fmt"
	"time"

	"github.com/Vaayne/aienvoy/pkg/llm"
)

// Answer represents the response from the Bard AI service. It contains
// the generated text content, conversation ID, response ID, factuality queries,
// original text query, and any choices provided.
type Answer struct {
	Content           string
	ConversationID    string
	ResponseID        string
	FactualityQueries []any
	TextQuery         string
	Choices           []Choice
	Links             []string
	Images            []any
	ProgramLang       string
	Code              string
	StatusCode        int
}

// Choice represents an alternative response option provided by Bard.
type Choice struct {
	// ID is a unique identifier for the choice.
	ID string
	// Content is the text content of the choice.
	Content string
}

func (a *Answer) ToChatCompletionResponse() llm.ChatCompletionResponse {
	return llm.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s:%s:%s", a.ConversationID, a.ResponseID, a.Choices[0].ID),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []llm.ChatCompletionChoice{
			{
				Index: 0,
				Message: llm.ChatCompletionMessage{
					Role:    "assistant",
					Content: a.Choices[0].Content,
				},
				FinishReason: "stop",
			},
		},
	}
}

func (a *Answer) ToChatCompletionStreamResponse() llm.ChatCompletionStreamResponse {
	return llm.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s:%s:%s", a.ConversationID, a.ResponseID, a.Choices[0].ID),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []llm.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: llm.ChatCompletionStreamChoiceDelta{
					Role:    "assistant",
					Content: a.Choices[0].Content,
				},
				FinishReason: "stop",
			},
		},
	}
}
