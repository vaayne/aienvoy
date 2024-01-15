package aigateway

import (
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

type AiGateway struct {
	*llm.LLM
}

func New(config llm.Config, dao llm.Dao) (*AiGateway, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}
	return &AiGateway{
		LLM: llm.New(dao, client),
	}, nil
}
