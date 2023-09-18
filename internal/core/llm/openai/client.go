package openai

import (
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/Vaayne/aienvoy/internal/pkg/config"

	"github.com/sashabaranov/go-openai"
)

const (
	LLMTypeOpenAI      = "openai"
	LLMTypeAzureOpenAI = "azure-openai"
)

type llmClient struct {
	openai.Client
	config.LLMConfig
	valid *atomic.Bool
}

func (c *llmClient) String() string {
	return fmt.Sprintf("llmClient{type: %s, endpoint: %s, version: %s}", c.Type, c.ApiEndpoint, c.ApiVersion)
}

func (c *llmClient) isValid(model string) bool {
	if !c.valid.Load() {
		return false
	}
	if len(c.Models) == 0 {
		return true
	}
	for _, m := range c.Models {
		if m == model {
			return true
		}
	}
	return false
}

var (
	clientPoolMap = make(map[int32]*llmClient)
	clientPoolIdx atomic.Int32
)

func init() {
	// init client pool
	var idx int32 = 0
	for _, llmCfg := range config.GetConfig().LLMs {
		client := createClient(llmCfg)
		if client == nil {
			continue
		}
		clientPoolMap[idx] = client
		idx++
	}
}

func getClientByModel(model string) (client *llmClient) {
	defer func() {
		if client != nil {
			slog.Debug("get LLM client", "client", client.String())
		}
	}()
	for i := 0; i < len(clientPoolMap); i++ {
		idx := clientPoolIdx.Add(1)
		if idx >= int32(len(clientPoolMap)) {
			idx = 0
			clientPoolIdx.Store(idx)
		}
		client := clientPoolMap[idx]
		if client.isValid(model) {
			return client
		}
	}
	return nil
}

func createClient(cfg config.LLMConfig) *llmClient {
	var clientCfg openai.ClientConfig

	if cfg.Type == LLMTypeAzureOpenAI {
		clientCfg = openai.DefaultAzureConfig(cfg.ApiKey, cfg.ApiEndpoint)
		if cfg.ApiVersion != "" {
			clientCfg.APIVersion = cfg.ApiVersion
		}
	} else if cfg.Type == LLMTypeOpenAI {
		clientCfg = openai.DefaultConfig(cfg.ApiKey)
		if cfg.ApiEndpoint != "" {
			clientCfg.BaseURL = cfg.ApiEndpoint
		}
	} else {
		slog.Error("unknown LLM type", "type", cfg.Type)
		return nil
	}

	valid := &atomic.Bool{}
	valid.Store(true)

	return &llmClient{
		Client:    *openai.NewClientWithConfig(clientCfg),
		LLMConfig: cfg,
		valid:     valid,
	}
}
