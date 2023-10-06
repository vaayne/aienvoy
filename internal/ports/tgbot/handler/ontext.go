package handler

import (
	"fmt"
	"strings"

	tb "gopkg.in/telebot.v3"
)

const (
	CommandBard   = "bard"
	CommandRead   = "read"
	CommandClaude = "claude"
)

func OnText(c tb.Context) error {
	text := strings.TrimSpace(c.Text())
	if text == "" {
		return c.Reply("empty message")
	}

	// start new conversation
	if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandBard)) {
		return OnBardChat(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandClaude)) {
		return OnClaudeChat(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandRead)) {
		return OnReadEase(c)
	}

	// continue conversation
	llmCache, ok := getLLMConversationFromCache()
	if !ok {
		return c.Reply("Do not have any conversation, please use command to start a new conversation")
	}

	switch llmCache.Model {
	case bardModelName:
		bardConversationInfos := strings.Split(llmCache.Value, "-")
		return askBard(c, text, bardConversationInfos[0], bardConversationInfos[1], bardConversationInfos[2])
	case claudeModelName:
		return askClaude(c, llmCache.Value, text)
	}
	return c.Reply("Unsupport message")
}
