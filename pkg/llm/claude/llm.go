package claude

import (
	"github.com/Vaayne/aienvoy/pkg/llm"

	"github.com/aws/aws-sdk-go-v2/aws"
)

const (
	ModelClaudeV2            = "anthropic.claude-v2"
	ModelClaudeV1Dot3        = "anthropic.claude-v1"
	ModelClaudeInstantV1Dot2 = "anthropic.claude-instant-v1"
)

type Claude struct {
	*llm.LLM
}

func New(config aws.Config, dao llm.Dao) *Claude {
	return &Claude{
		LLM: llm.New(dao, NewClient(config)),
	}
}

func (c *Client) ListModels() []string {
	return ListModels()
}

func ListModels() []string {
	return []string{ModelClaudeV2, ModelClaudeV1Dot3, ModelClaudeInstantV1Dot2}
}
