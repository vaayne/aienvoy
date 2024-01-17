package llms

import (
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/llms"
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

func NewWithDao(model string, dao llm.Dao) (*llm.LLM, error) {
	return llms.NewWithDao(model, config.GetConfig().LLMs, dao)
}

func New(model string) (*llm.LLM, error) {
	return NewWithDao(model, llm.NewMemoryDao())
}
