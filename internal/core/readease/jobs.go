package readease

import (
	"context"
	"fmt"
	"sync"

	"aienvoy/internal/pkg/logger"
	"aienvoy/pkg/claudeweb"
	"aienvoy/pkg/hackernews"

	"github.com/pocketbase/pocketbase"
)

func ReadEasePeriodJob(app *pocketbase.PocketBase) ([]string, error) {
	ctx := context.Background()

	contents := make([]string, 0, 100)

	hn := hackernews.New()
	topStories, err := hn.GetTopStories(100)
	if err != nil {
		logger.SugaredLogger.Errorw("get top stories err", "err", err)
		return contents, fmt.Errorf("get top stories err: %w", err)
	}

	const numbJobs = 2
	wg := &sync.WaitGroup{}
	reader := NewReader(app)

	claude := claudeweb.DefaultClaudeWeb()
	for i := 0; i < len(topStories); i += numbJobs {
		for _, id := range topStories[i : i+numbJobs] {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				itemUrl := fmt.Sprintf("%s/item?id=%d", hackernews.HN_HOST, id)
				content, err := reader.readFromCache(ctx, itemUrl)
				if err != nil {
					logger.SugaredLogger.Infow("readFromCache err", "err", err)
				} else if content != "" {
					contents = append(contents, content)
					return
				}

				covId, prompt, err := reader.read(ctx, itemUrl)
				if err != nil {
					logger.SugaredLogger.Errorw("read err", "err", err)
					return
				}
				resp, err := claude.CreateChatMessage(covId, prompt)
				if err != nil {
					logger.SugaredLogger.Errorw("summaryArticle create chat message error", "err", err)
					return
				}
				logger.SugaredLogger.Infow("summaryArticle create chat message success", "resp", resp)
				contents = append(contents, resp.Completion)
			}(id)
		}
		wg.Wait()
	}
	return contents, nil
}
