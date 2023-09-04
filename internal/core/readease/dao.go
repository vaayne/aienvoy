package readease

import (
	"aienvoy/internal/pkg/dtoutils"
	"context"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

const TableReadeaseArticle = "readease_article"

type ReadeaseArticle struct {
	Id             string `json:"id,omitempty" mapstructure:"id,omitempty"`
	Url            string `json:"url,omitempty" mapstructure:"url,omitempty"`
	OriginalUrl    string `json:"original_url,omitempty" mapstructure:"original_url,omitempty"`
	Summary        string `json:"summary,omitempty" mapstructure:"summary,omitempty"`
	ViewCounts     int    `json:"view_counts,omitempty" mapstructure:"view_counts,omitempty"`
	Title          string `json:"title,omitempty" mapstructure:"title,omitempty"`
	Content        string `json:"content,omitempty" mapstructure:"content,omitempty"`
	LlmType        string `json:"llm_type,omitempty" mapstructure:"llm_type,omitempty"`
	LlmCovId       string `json:"llm_cov_id,omitempty" mapstructure:"llm_cov_id,omitempty"`
	IsReadeaseSent bool   `json:"is_readease_sent,omitempty" mapstructure:"is_readease_sent,omitempty"`
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
	err = dtoutils.FromRecord(record, article)
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

	if err := dtoutils.ToRecord(record, article); err != nil {
		return err
	}

	record.Set("view_counts", record.GetInt("view_counts")+1)
	return tx.SaveRecord(record)
}
