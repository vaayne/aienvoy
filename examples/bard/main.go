package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/bard"
	"github.com/Vaayne/aienvoy/pkg/cookiecloud"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	slog.SetDefault(logger)

	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	getCookie := func(key string) string {
		val, err := cc.GetCookie(".google.com", key)
		if err != nil {
			return ""
		}
		return val.Value
	}

	b, err := bard.NewBardClient(getCookie("__Secure-1PSID"), bard.WithCookies(map[string]string{
		"__Secure-1PSID":   getCookie("__Secure-1PSID"),
		"__Secure-1PSIDCC": getCookie("__Secure-1PSIDCC"),
		"__Secure-1PSIDTS": getCookie("__Secure-1PSIDTS"),
	}), bard.WithTimeout(60*time.Second))
	if err != nil {
		slog.Error("init bard error", "err", err)
		return
	}

	prompt := "hello, please summary latest news from hackernews"
	answer, err := b.Ask(prompt, "", "", "", 0)
	if err != nil {
		slog.Error("talk to bard error", "err", err)
		return
	}

	slog.Info("bard response", "answer", answer.Content, "choices", answer.Choices)

	prompt = "Please tell more"
	answer, err = b.Ask(prompt, answer.ConversationID, answer.ResponseID, answer.Choices[0].ID, 0)
	if err != nil {
		slog.Error("talk to bard error", "err", err)
		return
	}

	slog.Info("bard response", "answer", answer.Content, "choices", answer.Choices)
}
