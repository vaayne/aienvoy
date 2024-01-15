package awsbedrock

import (
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

type Claude struct {
	*llm.LLM
}

func New(config llm.Config, dao llm.Dao) (*Claude, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Claude{
		LLM: llm.New(dao, client),
	}, nil
}
