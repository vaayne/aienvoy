package openai

import (
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

type OpenAI *llm.LLM

func New(cfg llm.Config, dao llm.Dao) (OpenAI, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return llm.New(dao, client), nil
}
