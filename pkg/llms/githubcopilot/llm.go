package githubcopilot

import (
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

type GitHubCopilot *llm.LLM

func New(cfg llm.Config, dao llm.Dao) (GitHubCopilot, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return llm.New(dao, NewClient(cfg.ApiKey)), nil
}
