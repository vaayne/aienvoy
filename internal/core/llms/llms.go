package llms

import (
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/Vaayne/aienvoy/pkg/llm/aigateway"
	"github.com/Vaayne/aienvoy/pkg/llm/claude"
	llmconfig "github.com/Vaayne/aienvoy/pkg/llm/config"
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
			slog.Error("init openai client error", "err", err)
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
			if err := addClient(openai.New(cfg, dao)); err != nil {
				slog.Error("init openai client error", "err", err)
				continue
			}
		case llmconfig.LLMTypeTogether:
			if err := addClient(together.New(cfg, dao)); err != nil {
				slog.Error("init together client error", "err", err)
				continue
			}
		case llmconfig.LLMTypeAWSBedrock:
			if err := addClient(claude.New(cfg, dao)); err != nil {
				slog.Error("init claude client error", "err", err)
				continue
			}
		case llmconfig.LLMTypeAiGateway:
			if err := addClient(aigateway.New(cfg, dao)); err != nil {
				slog.Error("init aigateway client error", "err", err)
				continue
			}
		}
	}

	slog.Info("llm clients", "clients", modelLlmMapping)
	if len(modelLlmMapping) == 0 {
		log.Fatal("no llm clients found")
	}
}

func New(model string, dao llm.Dao) (llm.Interface, error) {
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

func DefaultLLM(model string) (llm.Interface, error) {
	return New(model, llm.NewMemoryDao())
}
