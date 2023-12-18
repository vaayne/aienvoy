package llms

import (
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/Vaayne/aienvoy/pkg/llm/client"
)

func NewWithDao(model string, dao llm.Dao) (llm.Interface, error) {
	return client.NewWithDao(model, config.GetConfig().LLMs, dao)
}

func New(model string) (llm.Interface, error) {
	return NewWithDao(model, llm.NewMemoryDao())
}
