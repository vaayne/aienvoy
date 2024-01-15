package googleai

import (
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

type GoogleAI struct {
	*llm.LLM
}

func New(cfg llm.Config, dao llm.Dao) (*GoogleAI, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &GoogleAI{
		LLM: llm.New(dao, NewClient(cfg.ApiKey)),
	}, nil
}
