package llm

import (
	"context"

	"github.com/Vaayne/aienvoy/internal/pkg/dtoutils"

	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

const TableLlmUsages = "llm_usages"

type LlmUsages struct {
	dtoutils.BaseModel
	ApiKey     string `json:"api_key,omitempty" mapstructure:"api_key,omitempty"`
	TokenUsage int64  `json:"token_usage,omitempty" mapstructure:"token_usage,omitempty"`
	UserId     string `json:"user_id,omitempty" mapstructure:"user_id,omitempty"`
	Model      string `json:"model,omitempty" mapstructure:"model,omitempty"`
}

func SaveLlmUsage(ctx context.Context, tx *daos.Dao, usage *LlmUsages) error {
	col, err := tx.FindCollectionByNameOrId(TableLlmUsages)
	if err != nil {
		return err
	}

	record := models.NewRecord(col)

	if err := dtoutils.ToRecord(record, usage); err != nil {
		return err
	}

	return tx.SaveRecord(record)
}
