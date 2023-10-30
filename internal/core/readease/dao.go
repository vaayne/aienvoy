package readease

import (
	"context"

	"github.com/Vaayne/aienvoy/internal/pkg/dtoutils"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

const TableReadeaseArticles = "readease_articles"

type Article struct {
	dtoutils.BaseModel
	Url            string `json:"url,omitempty" mapstructure:"url,omitempty"`
	OriginalUrl    string `json:"original_url,omitempty" mapstructure:"original_url,omitempty"`
	Summary        string `json:"summary,omitempty" mapstructure:"summary,omitempty"`
	ViewCount      int    `json:"view_count,omitempty" mapstructure:"view_count,omitempty"`
	Title          string `json:"title,omitempty" mapstructure:"title,omitempty"`
	Content        string `json:"content,omitempty" mapstructure:"content,omitempty"`
	LlmModel       string `json:"llm_model,omitempty" mapstructure:"llm_model,omitempty"`
	LlmCovId       string `json:"llm_cov_id,omitempty" mapstructure:"llm_cov_id,omitempty"`
	IsReadeaseSent bool   `json:"is_readease_sent,omitempty" mapstructure:"is_readease_sent,omitempty"`
}

func getArticleCollection(tx *daos.Dao) (*models.Collection, error) {
	return tx.FindCollectionByNameOrId(TableReadeaseArticles)
}

func findArticleByURL(ctx context.Context, tx *daos.Dao, url string) (*models.Record, error) {
	records, err := tx.FindRecordsByExpr(TableReadeaseArticles, dbx.NewExp("url={:url} or original_url={:url}", dbx.Params{"url": url}))
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return records[0], nil
}

func GetArticleByUrl(ctx context.Context, tx *daos.Dao, url string) (*Article, error) {
	record, err := findArticleByURL(ctx, tx, url)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}

	var article Article
	err = dtoutils.FromRecord(record, &article)
	return &article, err
}

func UpsertArticle(ctx context.Context, tx *daos.Dao, article *Article) error {
	record, err := findArticleByURL(ctx, tx, article.Url)
	if err != nil {
		return err
	}
	// if record not found, create a new one
	if record == nil {
		col, err := getArticleCollection(tx)
		if err != nil {
			return err
		}
		record = models.NewRecord(col)
	}

	if err := dtoutils.ToRecord(record, article); err != nil {
		return err
	}

	record.Set("view_count", record.GetInt("view_count")+1)
	return tx.SaveRecord(record)
}
