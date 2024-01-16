package githubcopilot

import "github.com/Vaayne/aienvoy/pkg/llms/llm"

type Request struct {
	Model         string                      `json:"model"`
	Messages      []llm.ChatCompletionMessage `json:"messages"`
	Temperature   float32                     `json:"temperature,omitempty"`
	TopP          float32                     `json:"top_p,omitempty"`
	N             int                         `json:"n,omitempty"`
	Stream        bool                        `json:"stream,omitempty"`
	Intent        bool                        `json:"intent"`
	OneTimeReturn bool                        `json:"one_time_return"`
	MaxTokens     int                         `json:"max_tokens,omitempty"`
}

func (r *Request) FromChatCompletionRequest(req llm.ChatCompletionRequest) {
	r.Model = req.ModelId()
	r.Messages = req.Messages
	r.Temperature = 0.5
	r.TopP = 1
	r.N = 1
	r.Stream = true
	r.Intent = true
	r.OneTimeReturn = true
	r.MaxTokens = 2048
}
