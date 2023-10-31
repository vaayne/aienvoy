package migrations

import (
	"log/slog"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

const tableNameMessages = "conversation_messages"

func init() {
	m.Register(func(db dbx.Builder) error {
		collection := &models.Collection{
			Name: tableNameMessages,
			Type: models.CollectionTypeBase,
			Indexes: types.JsonArray[string]{
				"CREATE INDEX idx_conversation_id ON midjourney_jobs (conversation_id, created, updated)",
				"CREATE INDEX idx_origin_message_id ON midjourney_jobs (origin_message_id, created)",
				"CREATE INDEX idx_updated ON midjourney_jobs (updated, model)",
			},
			Schema: schema.NewSchema(&schema.SchemaField{
				Name:     "user_id",
				Type:     schema.FieldTypeText,
				Required: true,
			}, &schema.SchemaField{
				Name:     "conversation_id",
				Type:     schema.FieldTypeText,
				Required: true,
			}, &schema.SchemaField{
				Name:     "origin_message_id",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "model",
				Type:     schema.FieldTypeText,
				Required: true,
			}, &schema.SchemaField{
				Name:     "prompt",
				Type:     schema.FieldTypeText,
				Required: true,
			}, &schema.SchemaField{
				Name:     "completion",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "max_tokens",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}, &schema.SchemaField{
				Name:     "temperature",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}, &schema.SchemaField{
				Name:     "prompt_token",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}, &schema.SchemaField{
				Name:     "completion_token",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}, &schema.SchemaField{
				Name:     "description",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}),
		}
		if err := daos.New(db).SaveCollection(collection); err != nil {
			slog.Error("create table error", "err", err, "table", tableNameMessages)
			return err
		}
		slog.Info("create table success", "table", tableNameMessages)
		return nil
	}, func(db dbx.Builder) error {
		collection, err := daos.New(db).FindCollectionByNameOrId(tableNameMessages)
		if err != nil {
			return err
		}
		if err := daos.New(db).DeleteCollection(collection); err != nil {
			slog.Error("drop table error", "err", err, "table", tableNameMessages)
			return err
		}
		slog.Info("drop table success", "table", tableNameMessages)
		return nil
	})
}
