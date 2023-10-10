package handler

import (
	"fmt"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm/llmbard"
	tb "gopkg.in/telebot.v3"
)

// const welcomeMsg = "Welcome to Bard Chat! How may I assist you today? You can respond with any text message and Bard will promptly respond. Please note that if there is no response within five minutes, the current session will end."

const bardModelName = "Google Bard"

func OnBardChat(c tb.Context) error {
	text := strings.TrimSpace(c.Text()[1+len(CommandBard):])
	if text == "" {
		text = "hello"
	}
	return askBard(c, text, "", "", "")
}

func askBard(c tb.Context, prompt, conversationID, responseID, choiceID string) error {
	bard := llmbard.New()
	resp, err := bard.Ask(prompt, conversationID, responseID, choiceID, 0)
	if err != nil {
		return c.Reply(fmt.Sprintf("bard got an error, %s", err))
	}

	setLLMConversationToCache(LLMCache{
		Model:        bardModelName,
		Conversation: fmt.Sprintf("%s-%s-%s", resp.ConversationID, resp.ResponseID, resp.Choices[0].ID),
	})
	return c.Send(resp.Choices[0].Content)
}
