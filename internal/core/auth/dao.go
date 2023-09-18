package auth

import (
	"context"
	"fmt"

	"github.com/Vaayne/aienvoy/internal/pkg/dtoutils"

	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

const (
	TableApiKeys = "api_keys"
	ColumnApiKey = "api_key"
)

type ApiKey struct {
	dtoutils.BaseModel
	ApiKey    string   `json:"api_key,omitempty" mapstructure:"api_key,omitempty"`
	UserId    string   `json:"user_id,omitempty" mapstructure:"user_id,omitempty"`
	LlmModels []string `json:"llm_models,omitempty" mapstructure:"llm_models,omitempty"`
}

func FindAuthRecordByApiKey(ctx context.Context, tx *daos.Dao, apiKey string) (*models.Record, error) {
	record, err := tx.FindFirstRecordByData(TableApiKeys, ColumnApiKey, apiKey)
	if err != nil {
		return nil, err
	}

	if record == nil {
		return nil, fmt.Errorf("record not exist")
	}
	return record, nil
}
