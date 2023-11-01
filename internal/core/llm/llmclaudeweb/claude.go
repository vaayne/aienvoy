package llmclaudeweb

import (
	"log/slog"
	"time"

	"github.com/Vaayne/aienvoy/internal/pkg/cache"
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/claudeweb"
	"github.com/Vaayne/aienvoy/pkg/cookiecloud"
)

type ClaudeWeb struct {
	*claudeweb.ClaudeWeb
}

func New() *ClaudeWeb {
	client, _ := cache.CacheFunc(func(params ...any) (any, error) {
		return &ClaudeWeb{newClient()}, nil
	}, "claudeWebClientCacheKey", 2*time.Minute)
	return client.(*ClaudeWeb)
}

func newClient() *claudeweb.ClaudeWeb {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	sessionKey, err := cc.GetCookie("claude.ai", "sessionKey")
	if err != nil {
		slog.Error("get cookie error", "err", err)
		return nil
	}

	return claudeweb.New(sessionKey.Value)
}
