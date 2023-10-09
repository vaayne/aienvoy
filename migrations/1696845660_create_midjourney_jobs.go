package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		jsonData := `{
			"id": "o4cn6kgaxis96fz",
			"name": "midjourney_jobs",
			"type": "base",
			"system": false,
			"schema": [
				{
					"id": "hxb4uess",
					"name": "prompt",
					"type": "text",
					"system": false,
					"required": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"id": "bmvdktae",
					"name": "action",
					"type": "select",
					"system": false,
					"required": false,
					"options": {
						"maxSelect": 1,
						"values": [
							"Generate",
							"Upscale",
							"Variate",
							"Reset"
						]
					}
				},
				{
					"id": "4vhcphhr",
					"name": "status",
					"type": "select",
					"system": false,
					"required": false,
					"options": {
						"maxSelect": 1,
						"values": [
							"Pending",
							"Processing",
							"Completed",
							"Failed"
						]
					}
				},
				{
					"id": "fxbwzvwd",
					"name": "channel_id",
					"type": "number",
					"system": false,
					"required": false,
					"options": {
						"min": null,
						"max": null
					}
				},
				{
					"id": "wawxsgt0",
					"name": "message_image_idx",
					"type": "number",
					"system": false,
					"required": false,
					"options": {
						"min": null,
						"max": null
					}
				},
				{
					"id": "dythkxuk",
					"name": "message_id",
					"type": "text",
					"system": false,
					"required": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"id": "uarkaa0m",
					"name": "message_hash",
					"type": "text",
					"system": false,
					"required": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"id": "nibeppup",
					"name": "image_name",
					"type": "text",
					"system": false,
					"required": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"id": "g7fgixmo",
					"name": "image_url",
					"type": "url",
					"system": false,
					"required": false,
					"options": {
						"exceptDomains": null,
						"onlyDomains": null
					}
				},
				{
					"id": "5el3p8t5",
					"name": "image_content_type",
					"type": "text",
					"system": false,
					"required": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"id": "t2ylfash",
					"name": "image_size",
					"type": "number",
					"system": false,
					"required": false,
					"options": {
						"min": null,
						"max": null
					}
				},
				{
					"id": "dwboojet",
					"name": "image_height",
					"type": "number",
					"system": false,
					"required": false,
					"options": {
						"min": null,
						"max": null
					}
				},
				{
					"id": "so7wdfcr",
					"name": "image_width",
					"type": "number",
					"system": false,
					"required": false,
					"options": {
						"min": null,
						"max": null
					}
				}
			],
			"indexes": [],
			"listRule": null,
			"viewRule": null,
			"createRule": null,
			"updateRule": null,
			"deleteRule": null,
			"options": {}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("o4cn6kgaxis96fz")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}
