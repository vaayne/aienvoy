package phind

import "github.com/sashabaranov/go-openai"

type Request struct {
	UserInput      string `json:"userInput"`
	Messages       any    `json:"messages"`
	PinnedMessages any    `json:"pinnedMessages"`
	AnonUserID     string `json:"anonUserID"`
}

func (r *Request) FromOpenAIChatCompletionRequest(req *openai.ChatCompletionRequest) {
	r.UserInput = req.Messages[len(req.Messages)-1].Content
	r.Messages = req.Messages[:len(req.Messages)-1]
}
