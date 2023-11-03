package llm

import (
	"log/slog"
	"time"

	"github.com/Vaayne/aienvoy/pkg/llm/phind"

	"github.com/Vaayne/aienvoy/internal/pkg/cache"
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/cookiecloud"
)

func newPhind() *phind.Phind {
	client, _ := cache.CacheFunc(func(params ...any) (any, error) {
		return newPhindClient(), nil
	}, "claudeWebClientCacheKey", 2*time.Minute)
	return client.(*phind.Phind)
}

func newPhindClient() *phind.Phind {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	cookies, err := cc.GetHttpCookies("www.phind.com")
	if err != nil {
		slog.Error("get cookies error", "err", err, "domain", "www.phind.com")
		return nil
	}

	cookies1, err := cc.GetHttpCookies(".phind.com")
	if err != nil {
		slog.Error("get cookies error", "err", err, "domain", ".phind.com")
		return nil
	}
	cookies = append(cookies, cookies1...)
	return phind.New(cookies)
}
