package readease

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/hackernews"
	"github.com/Vaayne/aienvoy/pkg/llm/claude"

	"github.com/pocketbase/pocketbase"
	"golang.org/x/sync/semaphore"
)

func PeriodJob(app *pocketbase.PocketBase) ([]string, error) {
	ctx := context.Background()
	slog.InfoContext(ctx, "Start readease period job...")
	topStoiresCnt := config.GetConfig().ReadEase.TopStoriesCnt

	contents := make([]string, 0, topStoiresCnt)

	hn := hackernews.New()
	stories, err := hn.GetBestStories(topStoiresCnt)
	if err != nil {
		return contents, fmt.Errorf("get top stories err: %w", err)
	}
	slog.InfoContext(ctx, "success get hackernews top stories", "count", len(stories))

	reader := NewReader(app)
	maxWorkers := runtime.GOMAXPROCS(0)
	sem := semaphore.NewWeighted(int64(maxWorkers))
	for _, id := range stories {
		if err := sem.Acquire(ctx, 1); err != nil {
			slog.ErrorContext(ctx, "Failed to acquire semaphore: %v", err)
			break
		}

		go func(id int) {
			defer sem.Release(1)
			itemUrl := fmt.Sprintf("%s/item?id=%d", hackernews.HNHost, id)
			slog.InfoContext(ctx, "start read hackernews item", "url", itemUrl)
			// get from cache

			article, err := reader.Read(ctx, itemUrl, claude.ModelClaudeV1Dot3)
			if err != nil {
				slog.ErrorContext(ctx, "read artilce error", "err", err)
				return
			}

			slog.DebugContext(ctx, "artilce meta", "url", article.Url, "title", article.Title, "content", article.Content[:min(len(article.Content), 100)], "summary", article.Summary[:min(len(article.Summary), 100)], "is_sent", article.IsReadeaseSent)

			if article.IsReadeaseSent {
				slog.DebugContext(ctx, "already sent article", "url", itemUrl, "title", article.Title)
				return
			}
			if article.Summary != "" {
				contents = append(contents, article.Summary)
				article.IsReadeaseSent = true
				if err := UpsertArticle(ctx, app.Dao(), article); err != nil {
					slog.ErrorContext(ctx, "error update article send status", "err", err)
				}
				return
			}
		}(id)
	}
	// Acquire all of the tokens to wait for any remaining workers to finish.
	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		slog.InfoContext(ctx, "Failed to acquire semaphore: %v", err)
	}
	slog.InfoContext(ctx, "success read top hackernews", "count", len(contents))
	return contents, nil
}
