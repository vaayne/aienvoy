package aiclient

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"openai-dashboard/internal/pkg/config"
	"openai-dashboard/internal/pkg/logger"
	"sync"
	"sync/atomic"
)

var (
	once     sync.Once
	aiClient *AiClient
)

const (
	LLMTypeOpenAI      = "openai"
	LLMTypeAzureOpenAI = "azure-openai"
)

type clientPool struct {
	idx     atomic.Int32
	clients []*openai.Client
}

type clientPoolMap map[string]*clientPool

func (cp *clientPool) addClient(client *openai.Client) {
	cp.clients = append(cp.clients, client)
}

type AiClient struct {
	modelClientMapping clientPoolMap
}

func (ac *AiClient) getClientByModel(model string) *openai.Client {
	if pool, ok := ac.modelClientMapping[model]; ok {
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

//func init() {
//	NewAiClient()
//}

func NewAiClient() *AiClient {
	if aiClient == nil {
		once.Do(func() {
			aiClient = &AiClient{
				modelClientMapping: make(clientPoolMap),
			}
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
						pool, ok := aiClient.modelClientMapping[model]
						if !ok {
							pool = &clientPool{}
							aiClient.modelClientMapping[model] = pool
						}
						pool.addClient(client)
					}
				}
			}
			logger.SugaredLogger.Debugw("ai client created", "client", aiClient.modelClientMapping)
		})
	}
	return aiClient
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
