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
			"id": "ab44yzklhaz5cqm",
			"created": "2023-09-04 06:08:12.782Z",
			"updated": "2023-09-04 06:08:12.782Z",
			"name": "llm_usages",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "ernthdss",
					"name": "api_key",
					"type": "relation",
					"required": false,
					"unique": false,
					"options": {
						"collectionId": "09smh9yko40albu",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": []
					}
				},
				{
					"system": false,
					"id": "2ccobehn",
					"name": "token_usage",
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
					"id": "rs9ypj3h",
					"name": "user_id",
					"type": "relation",
					"required": false,
					"unique": false,
					"options": {
						"collectionId": "_pb_users_auth_",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": []
					}
				},
				{
					"system": false,
					"id": "7lje8voy",
					"name": "model",
					"type": "text",
					"required": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_cpTDBvk` + "`" + ` ON ` + "`" + `llm_usages` + "`" + ` (\n  ` + "`" + `api_key` + "`" + `,\n  ` + "`" + `model` + "`" + `\n)",
				"CREATE INDEX ` + "`" + `idx_A2AvQ86` + "`" + ` ON ` + "`" + `llm_usages` + "`" + ` (\n  ` + "`" + `user_id` + "`" + `,\n  ` + "`" + `model` + "`" + `\n)"
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
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("ab44yzklhaz5cqm")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}
