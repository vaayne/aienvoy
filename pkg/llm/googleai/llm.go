package googleai

import (
	"github.com/Vaayne/aienvoy/pkg/llm"
	llmconfig "github.com/Vaayne/aienvoy/pkg/llm/config"
)

type GoogleAI struct {
	*llm.LLM
}

func New(cfg llmconfig.Config, dao llm.Dao) (*GoogleAI, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &GoogleAI{
		LLM: llm.New(dao, NewClient(cfg.ApiKey)),
	}, nil
}
