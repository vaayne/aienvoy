package auth

import (
	"context"
	"fmt"
	"time"

	"aienvoy/internal/pkg/dtoutils"

	"github.com/pocketbase/pocketbase/daos"
)

const (
	TableApiKeys = "api_keys"
	ColumnApiKey = "api_key"
)

type ApiKey struct {
	Id        string    `json:"id,omitempty" mapstructure:"id,omitempty"`
	Created   time.Time `json:"created,omitempty" mapstructure:"created,omitempty"`
	Updated   time.Time `json:"updated,omitempty" mapstructure:"updated,omitempty"`
	ApiKey    string    `json:"api_key,omitempty" mapstructure:"api_key,omitempty"`
	UserId    string    `json:"user_id,omitempty" mapstructure:"user_id,omitempty"`
	LlmModels []string  `json:"llm_models,omitempty" mapstructure:"llm_models,omitempty"`
}

func FindAuthRecordByApiKey(ctx context.Context, tx *daos.Dao, apiKey string) (*ApiKey, error) {
	record, err := tx.FindFirstRecordByData(TableApiKeys, ColumnApiKey, apiKey)
	if err != nil {
		return nil, err
	}

	if record == nil {
		return nil, fmt.Errorf("record not exist")
	}

	var key *ApiKey

	if err := dtoutils.FromRecord(record, key); err != nil {
		return nil, fmt.Errorf("error build apiKey %w", err)
	}

	return key, nil
}
