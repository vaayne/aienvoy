package dao

import (
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

const TableReadeaseArticle = "readease_article"

/*
	{
	  "id": "RECORD_ID",
	  "collectionId": "84flaib8zkqr634",
	  "collectionName": "readease_article",
	  "created": "2022-01-01 01:00:00.123Z",
	  "updated": "2022-01-01 23:59:59.456Z",
	  "url": "https://example.com",
	  "summary": "test",
	  "view_counts": 123,
	  "title": "test",
	  "content": "test"
	}
*/
type ReadeaseArticle struct {
	Id         string `json:"id"`
	Url        string `json:"url"`
	Summary    string `json:"summary"`
	ViewCounts int    `json:"view_counts"`
	Title      string `json:"title"`
	Content    string `json:"content"`
}

func getReadeaseArticleCollection(tx *daos.Dao) (*models.Collection, error) {
	return tx.FindCollectionByNameOrId(TableReadeaseArticle)
}

// GetReadeaseArticleByUrl returns a ReadeaseArticle by url
func GetReadeaseArticleByUrl(tx *daos.Dao, url string) (*ReadeaseArticle, error) {
	record, err := tx.FindFirstRecordByData(TableReadeaseArticle, "url", url)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	return &ReadeaseArticle{
		Id:         record.Id,
		Url:        record.GetString("url"),
		Summary:    record.GetString("summary"),
		ViewCounts: record.GetInt("view_counts"),
		Title:      record.GetString("title"),
		Content:    record.GetString("content"),
	}, nil
}

// CreateReadeaseArticle creates a ReadeaseArticle
func CreateReadeaseArticle(tx *daos.Dao, article *ReadeaseArticle) error {
	col, err := getReadeaseArticleCollection(tx)
	if err != nil {
		return err
	}
	record := models.NewRecord(col)
	record.Set("url", article.Url)
	record.Set("summary", article.Summary)
	record.Set("view_counts", article.ViewCounts)
	record.Set("title", article.Title)
	record.Set("content", article.Content)
	return tx.SaveRecord(record)
}
