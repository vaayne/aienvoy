package llms

import (
	"fmt"
	"testing"

	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

func TestSplitModel(t *testing.T) {
	modelLlmMapping = map[string]llm.Interface{
		llm.LLMTypeOpenAI.String(): nil,
	}

	tests := []struct {
		name  string
		model string
		want  string
	}{
		{
			name:  "Single word model",
			model: "model1",
			want:  "model1",
		},
		{
			name:  "Model with slash but no provider mapping",
			model: "provider1/model1",
			want:  "",
		},
		{
			name:  "Model with slash and provider mapping",
			model: fmt.Sprintf("%s/model1/model2", llm.LLMTypeOpenAI.String()),
			want:  "model1/model2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, got := splitModel(tt.model); got != tt.want {
				t.Errorf("splitModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
