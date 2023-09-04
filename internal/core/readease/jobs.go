package readease

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"aienvoy/pkg/hackernews"

	"github.com/pocketbase/pocketbase"
	"golang.org/x/sync/semaphore"
)

const NumOfTopStoires = 10

func ReadEasePeriodJob(app *pocketbase.PocketBase) ([]string, error) {
	slog.Info("Start readease period job...")
	ctx := context.Background()

	contents := make([]string, 0, NumOfTopStoires)

	hn := hackernews.New()
	stories, err := hn.GetBestStories(NumOfTopStoires)
	if err != nil {
		return contents, fmt.Errorf("get top stories err: %w", err)
	}
	slog.Info("success get hackernews top stories", "count", len(stories))

	reader := NewReader(app)
	maxWorkers := runtime.GOMAXPROCS(0)
	sem := semaphore.NewWeighted(int64(maxWorkers))
	for _, id := range stories {
		if err := sem.Acquire(ctx, 1); err != nil {
			slog.Error("Failed to acquire semaphore: %v", err)
			break
		}

		go func(id int) {
			defer sem.Release(1)
			itemUrl := fmt.Sprintf("%s/item?id=%d", hackernews.HN_HOST, id)
			slog.Info("start read hackernews item", "url", itemUrl)
			// get from cache

			article, err := reader.Read(ctx, itemUrl)
			if err != nil {
				slog.Error("read artilce error", "err", err)
				return
			}

			slog.Debug("artilce meta", "url", article.Url, "title", article.Title, "content", article.Content[:min(len(article.Content), 100)], "summary", article.Summary[:min(len(article.Summary), 100)], "is_sent", article.IsReadeaseSent)

			if article.IsReadeaseSent {
				slog.Debug("already sent article", "url", itemUrl, "title", article.Title)
				return
			}
			if article.Summary != "" {
				contents = append(contents, article.Summary)
				article.IsReadeaseSent = true
				if err := UpsertReadeaseArticle(ctx, app.Dao(), article); err != nil {
					slog.Error("error update article send status", "err", err)
				}
				return
			}
		}(id)
	}
	// Acquire all of the tokens to wait for any remaining workers to finish.
	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		slog.Info("Failed to acquire semaphore: %v", err)
	}
	slog.Info("success read top hackernews", "count", len(contents))
	return contents, nil
}
