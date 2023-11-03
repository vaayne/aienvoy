package phind

import (
	"github.com/Vaayne/aienvoy/pkg/llm"
)

type Request struct {
	UserInput      string `json:"userInput"`
	Messages       any    `json:"messages"`
	PinnedMessages any    `json:"pinnedMessages"`
	AnonUserID     string `json:"anonUserID"`
}

func (r *Request) FromChatCompletionRequest(req llm.ChatCompletionRequest) {
	r.UserInput = req.Messages[len(req.Messages)-1].Content
	r.Messages = req.Messages[:len(req.Messages)-1]
}
