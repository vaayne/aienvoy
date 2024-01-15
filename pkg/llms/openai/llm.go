package openai

import (
	"github.com/Vaayne/aienvoy/pkg/llms/llm"

	"github.com/sashabaranov/go-openai"
)

const (
	ModelGPT432K          = "gpt-4-32k"
	ModelGPT4             = "gpt-4"
	ModelGPT3Dot5Turbo16K = "gpt-3.5-turbo-16k"
	ModelGPT3Dot5Turbo    = "gpt-3.5-turbo"
)

var modelMappings = map[string]string{
	openai.GPT3Dot5Turbo: openai.GPT3Dot5Turbo1106,
	openai.GPT4:          openai.GPT4TurboPreview,
}

type OpenAI struct {
	*llm.LLM
}

func New(cfg llm.Config, dao llm.Dao) (*OpenAI, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &OpenAI{
		llm.New(dao, client),
	}, nil
}
