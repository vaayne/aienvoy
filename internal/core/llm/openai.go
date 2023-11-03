package llm

import (
	"log/slog"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/llm/openai"
	goopenai "github.com/sashabaranov/go-openai"
)

const (
	llmTypeAzureOpenAI = "azure-openai"
	llmTypeOpenAI      = "openai"
)

// OpenAI is a client for AI services

func newOpenai() *openai.OpenAI {
	var clientCfg goopenai.ClientConfig
	cfg := config.GetConfig().LLMs[0]
	if cfg.Type == llmTypeAzureOpenAI {
		clientCfg = goopenai.DefaultAzureConfig(cfg.ApiKey, cfg.ApiEndpoint)
		if cfg.ApiVersion != "" {
			clientCfg.APIVersion = cfg.ApiVersion
		}
	} else if cfg.Type == llmTypeOpenAI {
		clientCfg = goopenai.DefaultConfig(cfg.ApiKey)
		if cfg.ApiEndpoint != "" {
			clientCfg.BaseURL = cfg.ApiEndpoint
		}
	} else {
		slog.Error("unknown LLM type", "type", cfg.Type)
		return nil
	}

	return openai.New(clientCfg)
}
