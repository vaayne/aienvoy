package openai

import (
	"context"
	"sync/atomic"

	"openai-dashboard/internal/pkg/config"
	"openai-dashboard/internal/pkg/logger"

	"github.com/sashabaranov/go-openai"
)

var modelClientMapping clientPoolMap = make(clientPoolMap)

const (
	LLMTypeOpenAI      = "openai"
	LLMTypeAzureOpenAI = "azure-openai"
)

type clientPool struct {
	idx     atomic.Int32
	clients []*openai.Client
}

type clientPoolMap map[string]*clientPool

func (cp *clientPool) add(client *openai.Client) {
	cp.clients = append(cp.clients, client)
}

func init() {
	newClientPool()
}

func getClientByModel(model string) *openai.Client {
	if pool, ok := modelClientMapping[model]; ok {
		idx := pool.idx.Add(1)
		if idx >= int32(len(pool.clients)) {
			idx = 0
			pool.idx.Store(idx)
		}
		client := pool.clients[idx]
		logger.SugaredLogger.Debugw("get client by model", "model", model, "client_idx", idx)
		return client
	}
	return nil
}

// NewAiClient creates a new ai client
func newClientPool() {
	for _, llmCfg := range config.GetConfig().LLMs {
		client := createClient(llmCfg)
		if client == nil {
			break
		}
		if llmCfg.Type == LLMTypeOpenAI && len(llmCfg.Models) == 0 {
			if models, err := client.ListModels(context.Background()); err == nil {
				for _, model := range models.Models {
					llmCfg.Models = append(llmCfg.Models, model.ID)
				}
			} else {
				logger.SugaredLogger.Errorw("failed to list models", "err", err, "type", llmCfg.Type, "endpoint", llmCfg.ApiEndpoint)
			}
		}
		logger.SugaredLogger.Debugw("llm models", "models", llmCfg.Models)
		if len(llmCfg.Models) > 0 {
			for _, model := range llmCfg.Models {
				pool, ok := modelClientMapping[model]
				if !ok {
					pool = &clientPool{}
					modelClientMapping[model] = pool
				}
				pool.add(client)
			}
		}
	}
	logger.SugaredLogger.Debugw("ai client created", "client", modelClientMapping)
}

func createClient(cfg config.LLMConfig) *openai.Client {
	var clientCfg openai.ClientConfig

	if cfg.Type == LLMTypeAzureOpenAI {
		clientCfg = openai.DefaultAzureConfig(cfg.ApiKey, cfg.ApiEndpoint)
	} else if cfg.Type == LLMTypeOpenAI {
		clientCfg = openai.DefaultConfig(cfg.ApiKey)
		if cfg.ApiEndpoint != "" {
			clientCfg.BaseURL = cfg.ApiEndpoint
		}
	} else {
		logger.SugaredLogger.Errorw("unknown LLM type", "type", cfg.Type)
		return nil
	}
	return openai.NewClientWithConfig(clientCfg)
}
