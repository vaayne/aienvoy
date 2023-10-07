package handler

import (
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/Vaayne/aienvoy/pkg/cache"
)

type LLMCache struct {
	Model        string                         // Model name
	Conversation string                         // Conversation ID info
	Messages     []openai.ChatCompletionMessage // history messages
}

const llmCacheKey = "telegramLLMCacheKey"

func setLLMConversationToCache(llmCache LLMCache) {
	c := cache.DefaultClient
	c.Set(llmCacheKey, &llmCache, 5*time.Minute)
}

func getLLMConversationFromCache() (*LLMCache, bool) {
	val, ok := cache.DefaultClient.Get(llmCacheKey)
	if ok {
		return val.(*LLMCache), ok
	}
	return nil, ok
}
