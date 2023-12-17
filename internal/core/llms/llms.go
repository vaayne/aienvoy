package llms

import (
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/Vaayne/aienvoy/pkg/llm/aigateway"
	"github.com/Vaayne/aienvoy/pkg/llm/awsbedrock"
	llmconfig "github.com/Vaayne/aienvoy/pkg/llm/config"
	"github.com/Vaayne/aienvoy/pkg/llm/googleai"
	"github.com/Vaayne/aienvoy/pkg/llm/openai"
	"github.com/Vaayne/aienvoy/pkg/llm/together"
)

var (
	modelLlmMapping = make(map[string]llm.Interface)
	once            sync.Once
)

func initModelMapping(dao llm.Dao) {
	addClient := func(cli llm.Interface, err error) error {
		if err != nil {
			return err
		}

		for _, model := range cli.ListModels() {
			modelLlmMapping[model] = cli
		}
		return nil
	}

	for _, cfg := range config.GetConfig().LLMs {
		switch cfg.LLMType {
		case llmconfig.LLMTypeOpenAI, llmconfig.LLMTypeAzureOpenAI:
			cli, err := openai.New(cfg, dao)
			if err := addClient(cli, err); err != nil {
				slog.Error("init openai client error", "err", err, "config", cfg)
				continue
			}
		case llmconfig.LLMTypeTogether:
			if err := addClient(together.New(cfg, dao)); err != nil {
				slog.Error("init together client error", "err", err, "config", cfg)
				continue
			}
		case llmconfig.LLMTypeGoogleAI:
			cli, err := googleai.New(cfg, dao)
			if err := addClient(cli, err); err != nil {
				slog.Error("init googleai client error", "err", err, "config", cfg)
				continue
			}
		case llmconfig.LLMTypeAWSBedrock:
			cli, err := awsbedrock.New(cfg, dao)
			if err := addClient(cli, err); err != nil {
				slog.Error("init aws bedrock client error", "err", err, "config", cfg)
				continue
			}
		case llmconfig.LLMTypeAiGateway:
			cli, err := aigateway.New(cfg, dao)
			if err := addClient(cli, err); err != nil {
				slog.Error("init aigateway client error", "err", err, "config", cfg)
				continue
			}
		}
	}

	// get all keys from modelLlmMapping
	models := make([]string, 0, len(modelLlmMapping))
	for model := range modelLlmMapping {
		models = append(models, model)
	}
	slog.Debug("llm clients support models", "models", models)

	if len(modelLlmMapping) == 0 {
		log.Fatal("no llm clients found")
	}
}

func NewWithDao(model string, dao llm.Dao) (llm.Interface, error) {
	once.Do(func() {
		initModelMapping(dao)
	})
	if model == "" {
		return nil, fmt.Errorf("model is empty")
	}

	cli, ok := modelLlmMapping[model]
	if !ok {
		return nil, fmt.Errorf("client for model %s not found", model)
	}
	return cli, nil
}

func New(model string) (llm.Interface, error) {
	return NewWithDao(model, llm.NewMemoryDao())
}
