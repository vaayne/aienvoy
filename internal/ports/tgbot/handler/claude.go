package handler

import (
	"fmt"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm/llmclaudeweb"
	tb "gopkg.in/telebot.v3"
)

const claudeModelName = "Claude Web"

func OnClaudeChat(c tb.Context) error {
	text := strings.TrimSpace(c.Data())
	if text == "" {
		text = "hello"
	}
	return askClaude(c, "", text)
}

func askClaude(c tb.Context, id, prompt string) error {
	claude := llmclaudeweb.New()

	if id == "" {
		cov, err := claude.CreateConversation(prompt[:min(10, len(prompt))])
		if err != nil {
			return c.Reply("Create new claude conversiton error: %s", err)
		}
		id = cov.UUID
	}

	resp, err := claude.CreateChatMessage(id, prompt)
	if err != nil {
		return c.Reply(fmt.Sprintf("claude got an error, %s", err))
	}
	setLLMConversationToCache(LLMCache{
		Model:        claudeModelName,
		Conversation: id,
	})
	return c.Send(resp.Completion)
}
