package handler

import (
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llms/bard"
	"github.com/Vaayne/aienvoy/pkg/llms/claudeweb"
	"github.com/Vaayne/aienvoy/pkg/llms/llm"

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

	model := ""
	prompt := text
	if text[0] == '/' {
		texts := strings.Split(text, " ")
		model = texts[0][1:]
		if len(texts) > 1 {
			prompt = strings.Join(texts[1:], " ")
		} else {
			prompt = "hello"
		}

		switch model {
		case CommandBard:
			model = bard.ModelBard
		case CommandRead:
			return OnReadEase(c)
		case CommandChatGPT35:
			model = llm.OAIModelGPT3Dot5Turbo
		case CommandChatGPT4:
			model = llm.OAIModelGPT4TurboPreview
		case CommandClaudeWeb:
			model = claudeweb.ModelClaude2
		case CommandClaudeV2:
			model = llm.BedrockModelClaudeV2
		case CommandClaudeV1:
			model = llm.BedrockModelClaudeV1
		case CommandClaudeInstant:
			model = llm.BedrockModelClaudeInstantV1
		case CommandImagine:
			return OnMidJourneyImagine(c)
		default:
			return c.Reply("Unsupported command!")
		}
	}

	// 1. create new conversation, no cache, model != ""
	// 2. create new conversation, use cache, model != ""
	// 3. use cache, model == ""
	llmCache, ok := getLLMConversationFromCache()
	if ok {
		if model == "" {
			model = llmCache.Model
		}
		return onLLMChat(c, llmCache.ConversationId, model, prompt)
	}
	if model != "" {
		return onLLMChat(c, "", model, prompt)
	}

	return c.Reply("Unsupported message")
}
