package llm

import (
	"log/slog"
	"time"

	"github.com/Vaayne/aienvoy/pkg/llm/claudeweb"

	"github.com/Vaayne/aienvoy/internal/pkg/cache"
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/cookiecloud"
)

func newClaudeWeb() *claudeweb.ClaudeWeb {
	client, _ := cache.CacheFunc(func(params ...any) (any, error) {
		return newClaudeWebClient(), nil
	}, "claudeWebClientCacheKey", 2*time.Minute)
	return client.(*claudeweb.ClaudeWeb)
}

func newClaudeWebClient() *claudeweb.ClaudeWeb {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	sessionKey, err := cc.GetCookie("claude.ai", "sessionKey")
	if err != nil {
		slog.Error("get cookie error", "err", err)
		return nil
	}

	return claudeweb.New(sessionKey.Value)
}
