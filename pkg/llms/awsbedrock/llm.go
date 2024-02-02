package awsbedrock

import (
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

type Claude *llm.LLM

func New(config llm.Config, dao llm.Dao) (Claude, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}
	return llm.New(dao, client), nil
}
