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
	topStories, err := hn.GetTopStories(NumOfTopStoires)
	if err != nil {
		return contents, fmt.Errorf("get top stories err: %w", err)
	}
	slog.Info("success get hackernews top stories", "count", len(topStories))

	reader := NewReader(app)
	maxWorkers := runtime.GOMAXPROCS(0)
	sem := semaphore.NewWeighted(int64(maxWorkers))
	for _, id := range topStories {
		if err := sem.Acquire(ctx, 1); err != nil {
			slog.Error("Failed to acquire semaphore: %v", err)
			break
		}

		go func(id int) {
			defer sem.Release(1)
			itemUrl := fmt.Sprintf("%s/item?id=%d", hackernews.HN_HOST, id)
			slog.Info("start read hackernews item", "url", itemUrl)
			// get from cache
			summary, err := reader.Read(ctx, itemUrl)
			if err != nil {
				slog.Error("error read article by url", "err", err, "url", itemUrl)
				return
			}
			contents = append(contents, fmt.Sprintf("%s\n\n原文: %s", summary, itemUrl))
		}(id)
	}
	// Acquire all of the tokens to wait for any remaining workers to finish.
	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		slog.Info("Failed to acquire semaphore: %v", err)
	}
	slog.Info("success read top hackernews", "count", len(contents))
	return contents, nil
}
