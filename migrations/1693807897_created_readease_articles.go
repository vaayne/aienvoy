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
			"id": "g9pcikatzc29szh",
			"created": "2023-09-04 06:11:37.000Z",
			"updated": "2023-09-04 06:11:37.000Z",
			"name": "readease_articles",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "0df7i0ur",
					"name": "url",
					"type": "url",
					"required": false,
					"unique": false,
					"options": {
						"exceptDomains": null,
						"onlyDomains": null
					}
				},
				{
					"system": false,
					"id": "57bxcxwe",
					"name": "original_url",
					"type": "url",
					"required": false,
					"unique": false,
					"options": {
						"exceptDomains": null,
						"onlyDomains": null
					}
				},
				{
					"system": false,
					"id": "dngzil75",
					"name": "title",
					"type": "text",
					"required": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "pqthaqno",
					"name": "content",
					"type": "text",
					"required": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "zq67hodf",
					"name": "summary",
					"type": "text",
					"required": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "fyyuquhj",
					"name": "view_count",
					"type": "number",
					"required": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null
					}
				},
				{
					"system": false,
					"id": "ichptcm7",
					"name": "llm_model",
					"type": "text",
					"required": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "zmi2i2t6",
					"name": "llm_cov_id",
					"type": "text",
					"required": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "ifmekqjo",
					"name": "is_readease_sent",
					"type": "bool",
					"required": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_GsXzUhy` + "`" + ` ON ` + "`" + `readease_articles` + "`" + ` (` + "`" + `url` + "`" + `)",
				"CREATE UNIQUE INDEX ` + "`" + `idx_IOVVbjs` + "`" + ` ON ` + "`" + `readease_articles` + "`" + ` (` + "`" + `original_url` + "`" + `)"
			],
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
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("g9pcikatzc29szh")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}
