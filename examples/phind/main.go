package main

import (
	"context"
	"github.com/Vaayne/aienvoy/pkg/phind"
	"github.com/sashabaranov/go-openai"
	"log/slog"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/cookiecloud"
)

func main() {
	ctx := context.Background()
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	cookies, err := cc.GetHttpCookies("www.phind.com")
	if err != nil {
		slog.Error("get cookies error", "err", err, "domain", "www.phind.com")
		return
	}

	cookies1, err := cc.GetHttpCookies(".phind.com")
	if err != nil {
		slog.Error("get cookies error", "err", err, "domain", ".phind.com")
		return
	}
	cookies = append(cookies, cookies1...)

	p := phind.New(cookies)
	resp, err := p.CreateCompletion(ctx, &openai.CompletionRequest{
		Prompt: "I want to find a job",
	})
	if err != nil {
		slog.Error("chat with claude error", "err", err)
		return
	}

	slog.Info("phind response message", "message", resp)
}
