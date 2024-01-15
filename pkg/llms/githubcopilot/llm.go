package githubcopilot

import (
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

type GitHubCopilot struct {
	*llm.LLM
}

func New(cfg llm.Config, dao llm.Dao) (*GitHubCopilot, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &GitHubCopilot{
		LLM: llm.New(dao, NewClient(cfg.ApiKey)),
	}, nil
}
