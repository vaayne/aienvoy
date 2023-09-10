package llmclaudeweb

import (
	"log/slog"
	"time"

	"aienvoy/internal/pkg/cache"
	"aienvoy/internal/pkg/config"
	"aienvoy/pkg/claudeweb"
	"aienvoy/pkg/cookiecloud"
)

func New() *claudeweb.ClaudeWeb {
	client, _ := cache.CacheFunc(func(params ...any) (any, error) {
		return newClient(), nil
	}, "claudeWebClientCacheKey", 2*time.Minute)
	return client.(*claudeweb.ClaudeWeb)
}

func newClient() *claudeweb.ClaudeWeb {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	sessionKey, err := cc.GetCookie("claude.ai", "sessionKey")
	if err != nil {
		slog.Error("get cookie error", "err", err)
		return nil
	}

	return claudeweb.NewClaudeWeb(sessionKey.Value)
}
