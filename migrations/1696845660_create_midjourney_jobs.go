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

const tableNameMidjourneyJobs = "midjourney_jobs"

func init() {
	m.Register(func(db dbx.Builder) error {
		collection := &models.Collection{
			Name: tableNameMidjourneyJobs,
			Type: models.CollectionTypeBase,
			Indexes: types.JsonArray[string]{
				"CREATE INDEX idx_channel_status ON midjourney_jobs (channel_id, status)",
			},
			Schema: schema.NewSchema(&schema.SchemaField{
				Name:     "prompt",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "action",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "status",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "channel_id",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}, &schema.SchemaField{
				Name:     "message_image_idx",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}, &schema.SchemaField{
				Name:     "message_id",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "message_hash",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "message_content",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "image_name",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "image_url",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "image_content_type",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "telegram_file_id",
				Type:     schema.FieldTypeText,
				Required: false,
			}, &schema.SchemaField{
				Name:     "image_size",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}, &schema.SchemaField{
				Name:     "image_height",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}, &schema.SchemaField{
				Name:     "image_width",
				Type:     schema.FieldTypeNumber,
				Required: false,
			}),
		}
		if err := daos.New(db).SaveCollection(collection); err != nil {
			slog.Error("createTableMidjourneyJob error", "err", err)
			return err
		}

		//if _, err := db.NewQuery(createTableMidjourneyJob).Execute(); err != nil {
		//	slog.Error("createTableMidjourneyJob error", "err", err)
		//	return err
		//}
		//
		//if _, err := db.NewQuery(createIndexForTableMidjourneyJob).Execute(); err != nil {
		//	slog.Error("createIndexForTableMidjourneyJob error", "err", err)
		//	return err
		//}

		return nil
	}, func(db dbx.Builder) error {
		collection, err := daos.New(db).FindCollectionByNameOrId(tableNameMidjourneyJobs)
		if err != nil {
			return err
		}
		if err := daos.New(db).DeleteCollection(collection); err != nil {
			slog.Error("deleteTableMidjourneyJob error", "err", err)
			return err
		}

		//if _, err := db.NewQuery(dropTableMidjourneyJob).Execute(); err != nil {
		//	slog.Error("dropTableMidjourneyJob error", "err", err)
		//	return err
		//}
		return nil
	})
}
