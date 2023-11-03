package handler

import (
	"fmt"
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llm/bard"
	"github.com/Vaayne/aienvoy/pkg/llm/claude"
	"github.com/Vaayne/aienvoy/pkg/llm/claudeweb"
	"github.com/Vaayne/aienvoy/pkg/llm/openai"

	tb "gopkg.in/telebot.v3"
)

const (
	CommandBard          = "bard"
	CommandRead          = "read"
	CommandChatGPT35     = "gpt35"
	CommandChatGPT4      = "gpt4"
	CommandClaudeWeb     = "claude_web"
	CommandClaudeV2      = "claude_v2"
	CommandClaudeV1      = "claude_v1"
	CommandClaudeInstant = "claude_instant"
	CommandImagine       = "imagine"
)

func OnText(c tb.Context) error {
	text := strings.TrimSpace(c.Text())
	if text == "" {
		return c.Reply("empty message")
	}

	// start new conversation
	if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandBard)) {
		return onLLMChat(c, bard.ModelBard)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandRead)) {
		return OnReadEase(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandChatGPT35)) {
		return onLLMChat(c, openai.ModelGPT3Dot5Turbo)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandChatGPT4)) {
		return onLLMChat(c, openai.ModelGPT4)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandClaudeWeb)) {
		return onLLMChat(c, claudeweb.ModelClaudeWeb)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandClaudeV2)) {
		return onLLMChat(c, claude.ModelClaudeV2)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandClaudeV1)) {
		return onLLMChat(c, claude.ModelClaudeV1Dot3)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandClaudeInstant)) {
		return onLLMChat(c, claude.ModelClaudeInstantV1Dot2)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandImagine)) {
		return OnMidJourneyImagine(c)
	}
	return c.Reply("Unsupported message")
}
