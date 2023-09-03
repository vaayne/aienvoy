package readease

import (
	"context"
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

const TableReadeaseArticle = "readease_article"

type ReadeaseArticle struct {
	Id             string `json:"id"`
	Url            string `json:"url"`
	OriginalUrl    string `json:"original_url"`
	Summary        string `json:"summary"`
	ViewCounts     int    `json:"view_counts"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	LlmType        string `json:"llm_type"`
	LlmCovId       string `json:"llm_cov_id"`
	IsReadeaseSent bool   `json:"is_readease_sent"`
}

func (a *ReadeaseArticle) FromRecord(r *models.Record) error {
	jsonData, err := r.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, a)
}

func (a *ReadeaseArticle) ToRecord(r *models.Record) error {
	jsonData, err := json.Marshal(a)
	if err != nil {
		return err
	}
	mapData := make(map[string]any)

	if err := json.Unmarshal(jsonData, &mapData); err != nil {
		return err
	}

	r.Load(mapData)
	return nil
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

	var article *ReadeaseArticle

	err = article.FromRecord(record)
	return article, err
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

	if err := article.ToRecord(record); err != nil {
		return err
	}

	record.Set("view_counts", record.GetInt("view_counts")+1)
	return tx.SaveRecord(record)
}
