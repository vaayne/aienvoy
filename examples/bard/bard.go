package main

import (
	"aienvoy/internal/pkg/config"
	"aienvoy/pkg/bard"
	"log/slog"
)

func main() {
	b, err := bard.NewBardClient(config.GetConfig().Bard.Token)
	if err != nil {
		slog.Error("init bard error", "err", err)
		return
	}

	prompt := "hello"
	answer, err := b.Ask(prompt, "", "", "", 0)
	if err != nil {
		slog.Error("talk to bard error", "err", err)
		return
	}

	slog.Info("bard response", "answer", answer)
}
