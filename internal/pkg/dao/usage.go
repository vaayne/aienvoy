package dao

import (
	"errors"
	"time"

	"aienvoy/internal/pkg/logger"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

const tableUsage = "usage"

type Usage struct {
	UserId   string    `json:"user_id"`
	ApiKey   string    `json:"api_key"`
	Usage    int       `json:"usage"`
	Model    string    `json:"model"`
	DateTime time.Time `json:"datetime"`
}

func (d *Dao) getUsageCollection(tx *daos.Dao) (*models.Collection, error) {
	return tx.FindCollectionByNameOrId(tableUsage)
}

func (d *Dao) CreateUsage(tx *daos.Dao, usage *Usage) error {
	records, err := tx.FindRecordsByExpr(
		tableUsage,
		dbx.HashExp{
			"user_id": usage.UserId,
			"api_key": usage.ApiKey,
			"model":   usage.Model,
		},
		dbx.NewExp("datetime>=datetime({:dt})", dbx.Params{"dt": usage.DateTime.Format(time.RFC3339)}),
	)
	if err != nil {
		return errors.New("find usage error: " + err.Error())
	}

	var record *models.Record
	if len(records) > 0 {
		record = records[0]
	} else {
		col, err := d.getUsageCollection(tx)
		if err != nil {
			return err
		}
		record = models.NewRecord(col)
		record.Set("usage", 0)
	}

	logger.SugaredLogger.Debug("save usage", "usage", usage, "record", record)
	record.Set("user_id", usage.UserId)
	record.Set("api_key", usage.ApiKey)
	record.Set("usage", usage.Usage+record.GetInt("usage"))
	record.Set("model", usage.Model)
	record.Set("datetime", usage.DateTime)

	return tx.SaveRecord(record)
}
