package handler

import (
	"fmt"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm/llmclaude"
	"github.com/sashabaranov/go-openai"

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
		return OnBardChat(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandRead)) {
		return OnReadEase(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandChatGPT35)) {
		return OnChatGPT35(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandChatGPT4)) {
		return OnChatGPT4(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandClaudeWeb)) {
		return OnClaudeChat(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandClaudeV2)) {
		return OnClaudeV2(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandClaudeV1)) {
		return OnClaudeV1Dot3(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandClaudeInstant)) {
		return OnClaudeInstant(c)
	} else if strings.HasPrefix(text, fmt.Sprintf("/%s", CommandImagine)) {
		return OnMidJourneyImagine(c)
	}

	// continue conversation
	llmCache, ok := getLLMConversationFromCache()
	if !ok {
		return c.Reply("Do not have any conversation, please use command to start a new conversation")
	}

	switch llmCache.Model {
	case bardModelName:
		bardConversationInfos := strings.Split(llmCache.Conversation, "-")
		return askBard(c, text, bardConversationInfos[0], bardConversationInfos[1], bardConversationInfos[2])
	case claudeModelName:
		return askClaude(c, llmCache.Conversation, text)
	case openai.GPT4, openai.GPT3Dot5Turbo, llmclaude.ModelClaudeV1Dot3, llmclaude.ModelClaudeV2, llmclaude.ModelClaudeInstantV1Dot2:
		return askLLM(c, llmCache.Conversation, llmCache.Model, text, llmCache.Messages)
	}
	return c.Reply("Unsupported message")
}
