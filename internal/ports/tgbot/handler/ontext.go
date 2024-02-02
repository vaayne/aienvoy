package handler

import (
	"fmt"
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llms/llm"

	tb "gopkg.in/telebot.v3"
)

const (
	CommandRead      = "read"
	CommandChatGPT35 = "gpt35"
	CommandChatGPT4  = "gpt4"
	CommandClaudeV2  = "claude_v2"
	CommandGemini    = "gemini"
	CommandImagine   = "imagine"
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
		case CommandRead:
			return OnReadEase(c)
		case CommandGemini:
			model = llm.DefaultGeminiModel
		case CommandChatGPT35:
			model = fmt.Sprintf("%s-%s/%s", llm.LLMTypeAiGateway, llm.AiGatewayProviderAzureOpenAI, llm.OAIModelGPT3Dot5Turbo)
		case CommandChatGPT4:
			model = fmt.Sprintf("%s-%s/%s", llm.LLMTypeAiGateway, llm.AiGatewayProviderAzureOpenAI, llm.OAIModelGPT4TurboPreview)
		case CommandClaudeV2:
			model = fmt.Sprintf("%s/%s", llm.LLMTypeAWSBedrock, llm.BedrockModelClaudeV2)
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
