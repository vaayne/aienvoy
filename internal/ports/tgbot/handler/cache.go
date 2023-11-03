package handler

import (
	"time"

	"github.com/Vaayne/aienvoy/pkg/cache"
)

type LLMCache struct {
	Model          string // Model name
	ConversationId string // Conversation ID info
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
