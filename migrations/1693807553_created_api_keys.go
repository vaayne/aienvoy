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
			"id": "09smh9yko40albu",
			"created": "2023-09-04 06:05:53.686Z",
			"updated": "2023-09-04 06:05:53.686Z",
			"name": "api_keys",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "vwew8itf",
					"name": "api_key",
					"type": "text",
					"required": true,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "mbosooki",
					"name": "user_id",
					"type": "relation",
					"required": false,
					"unique": false,
					"options": {
						"collectionId": "_pb_users_auth_",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": [
							"id"
						]
					}
				},
				{
					"system": false,
					"id": "iyyzkvrx",
					"name": "llm_models",
					"type": "json",
					"required": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_5APhVGQ` + "`" + ` ON ` + "`" + `api_keys` + "`" + ` (` + "`" + `api_key` + "`" + `)"
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

		collection, err := dao.FindCollectionByNameOrId("09smh9yko40albu")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}
