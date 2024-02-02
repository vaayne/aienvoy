package llms

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/Vaayne/aienvoy/pkg/llms/aigateway"
	"github.com/Vaayne/aienvoy/pkg/llms/anyscale"
	"github.com/Vaayne/aienvoy/pkg/llms/awsbedrock"
	"github.com/Vaayne/aienvoy/pkg/llms/llm"

	"github.com/Vaayne/aienvoy/pkg/llms/githubcopilot"
	"github.com/Vaayne/aienvoy/pkg/llms/googleai"
	"github.com/Vaayne/aienvoy/pkg/llms/openai"
	"github.com/Vaayne/aienvoy/pkg/llms/together"
)

var (
	modelLlmMapping map[string]*llm.LLM
	once            sync.Once
)

func getClient(cfg llm.Config, dao llm.Dao) (*llm.LLM, error) {
	switch cfg.LLMType {
	case llm.LLMTypeOpenAI, llm.LLMTypeAzureOpenAI, llm.LLMTypeOpenRouter:
		return openai.New(cfg, dao)
	case llm.LLMTypeTogether:
		return together.New(cfg, dao)
	case llm.LLMTypeAnyScale:
		return anyscale.New(cfg, dao)
	case llm.LLMTypeGoogleAI:
		return googleai.New(cfg, dao)
	case llm.LLMTypeAWSBedrock:
		return awsbedrock.New(cfg, dao)
	case llm.LLMTypeAiGateway:
		return aigateway.New(cfg, dao)
	case llm.LLMTypeGithubCopilot:
		return githubcopilot.New(cfg, dao)
	default:
		return nil, fmt.Errorf("client for type %s not found", cfg.LLMType)
	}
}

// initModelMapping initializes the modelLlmMapping map with the provided configurations.
// It creates a client for each configuration and maps it to the LLMType and Models in the configuration.
// If there's an error while creating a client, it logs the error and continues with the next configuration.
//
// Parameters:
// dao: An instance of llm.Dao which will be used to create the client.
// cfgs: A slice of llm.Config instances which contain the configurations for each client.
//
// Returns:
// This function doesn't return a value.
func initModelMapping(dao llm.Dao, cfgs []llm.Config) {
	// Initialize the modelLlmMapping map
	modelLlmMapping = make(map[string]*llm.LLM)

	// Iterate over the provided configurations
	for _, cfg := range cfgs {
		// Create a client for the current configuration
		cli, err := getClient(cfg, dao)
		if err != nil {
			// Log the error and continue with the next configuration if there's an error
			slog.Error("init client error", "err", err, "config", cfg)
			continue
		}

		// Map the client to the LLMType in the configuration
		modelLlmMapping[cfg.ID()] = cli

		// Map the client to each Model in the
		for _, model := range cfg.ListModels() {
			modelLlmMapping[model] = cli
		}
	}
}

func splitModel(model string) (string, string) {
	texts := strings.Split(model, "/")
	if len(texts) == 1 {
		return "", model
	}
	provider := texts[0]
	_, ok := modelLlmMapping[provider]
	if !ok {
		return "", ""
	}
	modelId := strings.Join(texts[1:], "/")
	return provider, modelId
}

func NewWithDao(model string, cfgs []llm.Config, dao llm.Dao) (*llm.LLM, error) {
	once.Do(func() {
		initModelMapping(dao, cfgs)
	})
	provider, modelId := splitModel(model)
	if modelId == "" {
		return nil, fmt.Errorf("model %s is not supported", model)
	}
	if provider != "" {
		return modelLlmMapping[provider], nil
	}
	cli, ok := modelLlmMapping[modelId]
	if ok {
		return cli, nil
	}
	return nil, fmt.Errorf("model %s is not supported", model)
}

func New(model string, cfgs []llm.Config) (*llm.LLM, error) {
	return NewWithDao(model, cfgs, llm.NewMemoryDao())
}
