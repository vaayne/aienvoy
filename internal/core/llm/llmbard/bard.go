package llmbard

import (
	"log/slog"
	"time"

	"aienvoy/internal/pkg/cache"
	"aienvoy/internal/pkg/config"
	"aienvoy/pkg/bard"
	"aienvoy/pkg/cookiecloud"
)

func New() *bard.BardClient {
	c, err := cache.CacheFunc(func(params ...any) (any, error) {
		return newBardClient()
	}, "bardClientCacheKey", 2*time.Minute)
	if err != nil {
		slog.Error("get bard client error", "err", err)
		return nil
	}
	return c.(*bard.BardClient)
}

func newBardClient() (*bard.BardClient, error) {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	getCookie := func(key string) string {
		val, err := cc.GetCookie(".google.com", key)
		if err != nil {
			slog.Error("get cookie error", "err", err)
			return ""
		}
		return val.Value
	}

	return bard.NewBardClient(getCookie("__Secure-1PSID"), bard.WithCookies(map[string]string{
		"__Secure-1PSID":   getCookie("__Secure-1PSID"),
		"__Secure-1PSIDCC": getCookie("__Secure-1PSIDCC"),
		"__Secure-1PSIDTS": getCookie("__Secure-1PSIDTS"),
	}))
}
