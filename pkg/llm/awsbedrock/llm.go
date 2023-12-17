package awsbedrock

import (
	"github.com/Vaayne/aienvoy/pkg/llm"
	llmconfig "github.com/Vaayne/aienvoy/pkg/llm/config"
)

type Claude struct {
	*llm.LLM
}

func New(config llmconfig.Config, dao llm.Dao) (*Claude, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Claude{
		LLM: llm.New(dao, client),
	}, nil
}
