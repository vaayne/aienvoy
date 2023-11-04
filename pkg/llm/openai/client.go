package openai

import (
	"github.com/sashabaranov/go-openai"
)

type Client struct {
	*openai.Client
}

func NewClient(cfg openai.ClientConfig) *Client {
	return &Client{
		Client: openai.NewClientWithConfig(cfg),
	}
}
