package aigateway

import (
	"github.com/Vaayne/aienvoy/pkg/llm"
	llmconfig "github.com/Vaayne/aienvoy/pkg/llm/config"
)

type AiGateway struct {
	*llm.LLM
}

func New(config llmconfig.Config, dao llm.Dao) (*AiGateway, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}
	return &AiGateway{
		LLM: llm.New(dao, client),
	}, nil
}
