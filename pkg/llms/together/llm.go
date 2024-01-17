package together

import (
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

type Together *llm.LLM

func New(cfg llm.Config, dao llm.Dao) (Together, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return llm.New(dao, client), nil
}
