package dao

import (
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

const tableApiKey = "api_keys"

const apiKeyColumnKey = "key"

type ApiKey struct {
	Key string `json:"key"`
}

func FindAuthRecordByApiKey(d *daos.Dao, apiKey string) (*models.Record, error) {
	return d.FindFirstRecordByData(tableApiKey, apiKeyColumnKey, apiKey)
}
