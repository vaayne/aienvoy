package llms

import (
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/Vaayne/aienvoy/pkg/llms/aigateway"
	"github.com/Vaayne/aienvoy/pkg/llms/awsbedrock"
	"github.com/Vaayne/aienvoy/pkg/llms/llm"

	"github.com/Vaayne/aienvoy/pkg/llms/githubcopilot"
	"github.com/Vaayne/aienvoy/pkg/llms/googleai"
	"github.com/Vaayne/aienvoy/pkg/llms/openai"
	"github.com/Vaayne/aienvoy/pkg/llms/together"
)

var (
	modelLlmMapping = make(map[string]llm.Interface)
	once            sync.Once
)

func initModelMapping(dao llm.Dao, cfgs []llm.Config) {
	addClient := func(cli llm.Interface, err error) error {
		if err != nil {
			return err
		}

		for _, model := range cli.ListModels() {
			modelLlmMapping[model] = cli
		}
		return nil
	}

	for _, cfg := range cfgs {
		switch cfg.LLMType {
		case llm.LLMTypeOpenAI, llm.LLMTypeAzureOpenAI, llm.LLMTypeOpenRouter:
			cli, err := openai.New(cfg, dao)
			if err := addClient(cli, err); err != nil {
				slog.Error("init openai client error", "err", err, "config", cfg)
				continue
			}
		case llm.LLMTypeTogether:
			if err := addClient(together.New(cfg, dao)); err != nil {
				slog.Error("init together client error", "err", err, "config", cfg)
				continue
			}
		case llm.LLMTypeGoogleAI:
			cli, err := googleai.New(cfg, dao)
			if err := addClient(cli, err); err != nil {
				slog.Error("init googleai client error", "err", err, "config", cfg)
				continue
			}
		case llm.LLMTypeAWSBedrock:
			cli, err := awsbedrock.New(cfg, dao)
			if err := addClient(cli, err); err != nil {
				slog.Error("init aws bedrock client error", "err", err, "config", cfg)
				continue
			}
		case llm.LLMTypeAiGateway:
			cli, err := aigateway.New(cfg, dao)
			if err := addClient(cli, err); err != nil {
				slog.Error("init aigateway client error", "err", err, "config", cfg)
				continue
			}
		case llm.LLMTypeGithubCopilot:
			cli, err := githubcopilot.New(cfg, dao)
			if err := addClient(cli, err); err != nil {
				slog.Error("init github copilot client error", "err", err, "config", cfg)
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

func NewWithDao(model string, cfgs []llm.Config, dao llm.Dao) (llm.Interface, error) {
	once.Do(func() {
		initModelMapping(dao, cfgs)
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

func New(model string, cfgs []llm.Config) (llm.Interface, error) {
	return NewWithDao(model, cfgs, llm.NewMemoryDao())
}
