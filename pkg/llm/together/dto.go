package together

import (
	"log/slog"
	"time"

	"github.com/Vaayne/aienvoy/pkg/llm"
)

// var defaultStops = []string{"</s>", "user:", "<|im_start|>", "<|im_end|>"}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type TogetherChatRequest struct {
	// Prompt            string    `json:"prompt"` // required
	Model             string    `json:"model"`      // required
	Messages          []Message `json:"messages"`   // required
	MaxTokens         int       `json:"max_tokens"` // required
	Stop              []string  `json:"stop,omitempty"`
	Temperature       float64   `json:"temperature,omitempty"`
	TopP              float64   `json:"top_p,omitempty"`
	TopK              int       `json:"top_k,omitempty"`
	RepetitionPenalty int       `json:"repetition_penalty,omitempty"`
	Logprobs          int       `json:"logprobs,omitempty"`
	Stream            bool      `json:"stream"`
}

func (r *TogetherChatRequest) FromChatCompletionRequest(req llm.ChatCompletionRequest) {
	r.Model = req.Model
	// r.Prompt = req.ToPrompt()
	for _, message := range req.Messages {
		r.Messages = append(r.Messages, Message{
			Role:    message.Role,
			Content: message.Content,
			Name:    message.Name,
		})
	}
	r.MaxTokens = req.MaxTokens
	r.Temperature = float64(req.Temperature)
	r.TopP = float64(req.TopP)
	r.Stream = req.Stream
	// req.Stop = defaultStops
	// if len(req.Stop) > 0 {
	// 	r.Stop = append(r.Stop, req.Stop...)
	// }

	slog.Info("from chat completion request", "tog req", r)
}

type TogetherChatResponseChoice struct {
	FinishReason string `json:"finish_reason"`
	Index        int    `json:"index"`
	Text         string `json:"text"`
}

// TogetherChatResponseChoice to llm.ChatCompletionChoice
func (c TogetherChatResponseChoice) ToChatCompletionChoice() llm.ChatCompletionChoice {
	return llm.ChatCompletionChoice{
		Message: llm.ChatCompletionMessage{
			Content: c.Text,
			Role:    llm.ChatMessageRoleAssistant,
		},
		Index:        c.Index,
		FinishReason: llm.FinishReason(c.FinishReason),
	}
}

func (c TogetherChatResponseChoice) ToChatCompletionStreamChoice() llm.ChatCompletionStreamChoice {
	return llm.ChatCompletionStreamChoice{
		Delta: llm.ChatCompletionStreamChoiceDelta{
			Content: c.Text,
			Role:    llm.ChatMessageRoleAssistant,
		},
		Index:        c.Index,
		FinishReason: llm.FinishReason(c.FinishReason),
	}
}

type TogetherChatResponse struct {
	Id      string                       `json:"id"`
	Choices []TogetherChatResponseChoice `json:"choices"`
	Created time.Time                    `json:"created"`
	Model   string                       `json:"model"`
	Object  string                       `json:"object"`
}

func (r TogetherChatResponse) ToChatCompletionResponse() llm.ChatCompletionResponse {
	resp := llm.ChatCompletionResponse{
		Choices: make([]llm.ChatCompletionChoice, len(r.Choices)),
	}
	for i, choice := range r.Choices {
		resp.Choices[i] = choice.ToChatCompletionChoice()
	}
	resp.Model = r.Model
	resp.ID = r.Id
	resp.Created = r.Created.Unix()
	return resp
}

func (r TogetherChatResponse) ToChatCompletionStreamResponse() llm.ChatCompletionStreamResponse {
	resp := llm.ChatCompletionStreamResponse{
		Choices: make([]llm.ChatCompletionStreamChoice, len(r.Choices)),
	}
	for i, choice := range r.Choices {
		// slog.Info("choice", "choice", choice, "stream choice", choice.ToChatCompletionStreamChoice())
		resp.Choices[i] = choice.ToChatCompletionStreamChoice()
	}
	resp.Model = r.Model
	resp.ID = r.Id
	resp.Created = r.Created.Unix()
	return resp
}
