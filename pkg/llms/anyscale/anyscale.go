package anyscale

import (
	"fmt"

	"github.com/Vaayne/aienvoy/pkg/llms/llm"
	"github.com/Vaayne/aienvoy/pkg/llms/openai"
)

const baseUrl = "https://api.endpoints.anyscale.com/v1"

var validModels = []string{
	"meta-llama/Llama-2-7b-chat-hf",
	"meta-llama/Llama-2-13b-chat-hf",
	"Meta-Llama/Llama-Guard-7b",
	"meta-llama/Llama-2-70b-chat-hf",
	"Open-Orca/Mistral-7B-OpenOrca",
	"codellama/CodeLlama-34b-Instruct-hf",
	"HuggingFaceH4/zephyr-7b-beta",
	"mistralai/Mistral-7B-Instruct-v0.1",
	"mistralai/Mixtral-8x7B-Instruct-v0.1",
	"thenlper/gte-large",
}

type AnyScale *llm.LLM

func New(cfg llm.Config, dao llm.Dao) (AnyScale, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return llm.New(dao, client), nil
}

type Client struct {
	*openai.Client
}

func NewClient(cfg llm.Config) (*Client, error) {
	if cfg.LLMType != llm.LLMTypeAnyScale {
		return nil, fmt.Errorf("invalid config, llmtype: %s", cfg.LLMType)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.BaseUrl == "" {
		cfg.BaseUrl = baseUrl
	}

	cli, err := openai.NewClient(cfg)

	return &Client{
		Client: cli,
	}, err
}

func (c *Client) ListModels() []string {
	return validModels
}
