package readease

import (
	"context"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

const TableReadeaseArticle = "readease_article"

type ReadeaseArticle struct {
	Id          string `json:"id"`
	Url         string `json:"url"`
	OriginalUrl string `json:"original_url"`
	Summary     string `json:"summary"`
	ViewCounts  int    `json:"view_counts"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	LlmType     string `json:"llm_type"`
	LlmCovId    string `json:"llm_cov_id"`
}

func getReadeaseArticleCollection(tx *daos.Dao) (*models.Collection, error) {
	return tx.FindCollectionByNameOrId(TableReadeaseArticle)
}

func findArticleByURL(ctx context.Context, tx *daos.Dao, url string) (*models.Record, error) {
	records, err := tx.FindRecordsByExpr(TableReadeaseArticle, dbx.NewExp("url={:url} or original_url={:url}", dbx.Params{"url": url}))
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return records[0], nil
}

func GetReadeaseArticleByUrl(ctx context.Context, tx *daos.Dao, url string) (*ReadeaseArticle, error) {
	record, err := findArticleByURL(ctx, tx, url)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	return &ReadeaseArticle{
		Id:          record.Id,
		Url:         record.GetString("url"),
		OriginalUrl: record.GetString("original_url"),
		Summary:     record.GetString("summary"),
		ViewCounts:  record.GetInt("view_counts"),
		Title:       record.GetString("title"),
		Content:     record.GetString("content"),
		LlmType:     record.GetString("llm_type"),
		LlmCovId:    record.GetString("llm_cov_id"),
	}, nil
}

func UpsertReadeaseArticle(ctx context.Context, tx *daos.Dao, article *ReadeaseArticle) error {
	record, err := findArticleByURL(ctx, tx, article.Url)
	if err != nil {
		return err
	}
	// if record not found, create a new one
	if record == nil {
		col, err := getReadeaseArticleCollection(tx)
		if err != nil {
			return err
		}
		record = models.NewRecord(col)
	}

	if article.Url != "" {
		record.Set("url", article.Url)
	}
	if article.OriginalUrl != "" {
		record.Set("original_url", article.OriginalUrl)
	}
	if article.Summary != "" {
		record.Set("summary", article.Summary)
	}

	record.Set("view_counts", record.GetInt("view_counts")+1)

	if article.Title != "" {
		record.Set("title", article.Title)
	}
	if article.Content != "" {
		record.Set("content", article.Content)
	}

	if article.LlmType != "" {
		record.Set("llm_type", article.LlmType)
	}
	if article.LlmCovId != "" {
		record.Set("llm_cov_id", article.LlmCovId)
	}
	return tx.SaveRecord(record)
}
