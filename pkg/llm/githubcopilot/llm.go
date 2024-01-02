package githubcopilot

import (
	"github.com/Vaayne/aienvoy/pkg/llm"
	llmconfig "github.com/Vaayne/aienvoy/pkg/llm/config"
)

type GitHubCopilot struct {
	*llm.LLM
}

func New(cfg llmconfig.Config, dao llm.Dao) (*GitHubCopilot, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &GitHubCopilot{
		LLM: llm.New(dao, NewClient(cfg.ApiKey)),
	}, nil
}
