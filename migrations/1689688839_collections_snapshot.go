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
		jsonData := `[
			{
				"id": "_pb_users_auth_",
				"created": "2023-07-10 01:00:23.289Z",
				"updated": "2023-07-12 14:33:37.044Z",
				"name": "users",
				"type": "auth",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "users_name",
						"name": "name",
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
						"id": "users_avatar",
						"name": "avatar",
						"type": "file",
						"required": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"maxSize": 5242880,
							"mimeTypes": [
								"image/jpeg",
								"image/png",
								"image/svg+xml",
								"image/gif",
								"image/webp"
							],
							"thumbs": null,
							"protected": false
						}
					}
				],
				"indexes": [],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {
					"allowEmailAuth": true,
					"allowOAuth2Auth": true,
					"allowUsernameAuth": true,
					"exceptEmailDomains": null,
					"manageRule": null,
					"minPasswordLength": 8,
					"onlyEmailDomains": null,
					"requireEmail": false
				}
			},
			{
				"id": "geqyaj3urpuwjos",
				"created": "2023-07-10 01:41:11.910Z",
				"updated": "2023-07-13 06:13:29.460Z",
				"name": "api_keys",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "5eb4cm2x",
						"name": "key",
						"type": "text",
						"required": true,
						"unique": false,
						"options": {
							"min": 20,
							"max": 51,
							"pattern": "^sk"
						}
					},
					{
						"system": false,
						"id": "hhnseadp",
						"name": "user_id",
						"type": "relation",
						"required": true,
						"unique": false,
						"options": {
							"collectionId": "_pb_users_auth_",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": []
						}
					}
				],
				"indexes": [
					"CREATE UNIQUE INDEX ` + "`" + `idx_5IYfscr` + "`" + ` ON ` + "`" + `api_keys` + "`" + ` (` + "`" + `key` + "`" + `)"
				],
				"listRule": "@request.auth.id != \"\" && user_id = @request.auth.id",
				"viewRule": "@request.auth.id != \"\" && user_id = @request.auth.id",
				"createRule": "@request.auth.id != \"\" && user_id = @request.auth.id",
				"updateRule": "@request.auth.id != \"\" && user_id = @request.auth.id",
				"deleteRule": "@request.auth.id != \"\" && user_id = @request.auth.id",
				"options": {}
			},
			{
				"id": "h6qviacalrn2c41",
				"created": "2023-07-18 13:55:55.608Z",
				"updated": "2023-07-18 13:58:08.250Z",
				"name": "usage_details",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "rrwcrk0z",
						"name": "user_id",
						"type": "relation",
						"required": false,
						"unique": false,
						"options": {
							"collectionId": "_pb_users_auth_",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": null,
							"displayFields": []
						}
					},
					{
						"system": false,
						"id": "cp4l67jb",
						"name": "api_key",
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
						"id": "pro8nmrc",
						"name": "usage",
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
						"id": "e1ecfycq",
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
					"CREATE INDEX ` + "`" + `idx_NnKuhxa` + "`" + ` ON ` + "`" + `usage_details` + "`" + ` (` + "`" + `user_id` + "`" + `)",
					"CREATE INDEX ` + "`" + `idx_Nwhk2uP` + "`" + ` ON ` + "`" + `usage_details` + "`" + ` (` + "`" + `api_key` + "`" + `)"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "dogizfwq41w36df",
				"created": "2023-07-18 13:59:20.666Z",
				"updated": "2023-07-18 13:59:33.889Z",
				"name": "usage",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "x5edsaac",
						"name": "api_key",
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
						"id": "jhk28lyo",
						"name": "usage",
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
						"id": "s9nyyyye",
						"name": "datetime",
						"type": "date",
						"required": false,
						"unique": false,
						"options": {
							"min": "",
							"max": ""
						}
					},
					{
						"system": false,
						"id": "wmdslqh7",
						"name": "user_id",
						"type": "relation",
						"required": false,
						"unique": false,
						"options": {
							"collectionId": "_pb_users_auth_",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": null,
							"displayFields": []
						}
					},
					{
						"system": false,
						"id": "b8hzufrx",
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
					"CREATE INDEX ` + "`" + `idx_Mmb9m3E` + "`" + ` ON ` + "`" + `usage` + "`" + ` (` + "`" + `user_id` + "`" + `)",
					"CREATE INDEX ` + "`" + `idx_OpD4pQ5` + "`" + ` ON ` + "`" + `usage` + "`" + ` (` + "`" + `api_key` + "`" + `)"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			}
		]`

		collections := []*models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collections); err != nil {
			return err
		}

		return daos.New(db).ImportCollections(collections, true, nil)
	}, func(db dbx.Builder) error {
		return nil
	})
}
