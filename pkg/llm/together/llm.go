package together

import (
	"github.com/Vaayne/aienvoy/pkg/llm"
	llmconfig "github.com/Vaayne/aienvoy/pkg/llm/config"
)

type Together struct {
	*llm.LLM
}

func New(cfg llmconfig.Config, dao llm.Dao) (*Together, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Together{
		llm.New(dao, client),
	}, nil
}
