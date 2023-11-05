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

const tableNameConversations = "conversations"

// id, created_on, updated_on, deleted, name, model, origin_id, description

func init() {
	m.Register(func(db dbx.Builder) error {
		collection := &models.Collection{
			Name: tableNameConversations,
			Type: models.CollectionTypeBase,
			Indexes: types.JsonArray[string]{
				"CREATE INDEX idx_created ON midjourney_jobs (created, model)",
				"CREATE INDEX idx_updated ON midjourney_jobs (updated, model)",
				"CREATE INDEX idx_user_id ON midjourney_jobs (updated, model)",
			},
			Schema: schema.NewSchema(&schema.SchemaField{
				Name:     "user_id",
				Type:     schema.FieldTypeText,
				Required: true,
			}, &schema.SchemaField{
				Name:     "name",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "model",
				Type:     schema.FieldTypeText,
				Required: true,
			}, &schema.SchemaField{
				Name:     "summary",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}, &schema.SchemaField{
				Name:     "extra_info",
				Type:     schema.FieldTypeText,
				Required: false,
			}),
		}
		if err := daos.New(db).SaveCollection(collection); err != nil {
			slog.Error("create table error", "err", err, "table", tableNameConversations)
			return err
		}
		slog.Info("create table success", "table", tableNameConversations)
		return nil
	}, func(db dbx.Builder) error {
		collection, err := daos.New(db).FindCollectionByNameOrId(tableNameConversations)
		if err != nil {
			return err
		}
		if err := daos.New(db).DeleteCollection(collection); err != nil {
			slog.Error("drop table error", "err", err, "table", tableNameConversations)
			return err
		}
		slog.Info("drop table success", "table", tableNameConversations)
		return nil
	})
}
