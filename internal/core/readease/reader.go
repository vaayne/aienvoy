package readease

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"strings"

	"aienvoy/internal/pkg/parser"
	"aienvoy/pkg/claudeweb"

	"github.com/pocketbase/pocketbase"
)

var promptTemplate = `
You are a college professor. You are reading an essay.
Here is the essay:
<essay>
{{.Content}}
</essay>
lease do the following job:
	1): Summarize this essay in a bullet point outline.
	2): Write down the main point of the essay and explain in details how the author prove the point.
	3): Write down the main idea of each paragraph .
	4): Write a Q&A notes, Ask 10 question about the essay and get answer from essay.
**Please response in Chinese and do not make other extra explaination and comment**. An example of response will be:
<example response>
## Summary
This essay is about how to write a summary

## Notes
- The author first introduce the importance of summary and then explain how to write a summary.
- ...

## Main Point
- The main point of this essay is how to write a summary.
- The author prove the point by ...

## Main Idea
- The main idea of each paragraph is ...
- ...

## Q&A
- Q: What is the main point of this essay?
- A: The main point of this essay is how to write a summary.
- Q: What is the main idea of the first paragraph?
- A: The main idea of the first paragraph is ...
</example response>
`

type ReadeaseReader struct {
	app *pocketbase.PocketBase
}

func NewReader(app *pocketbase.PocketBase) *ReadeaseReader {
	return &ReadeaseReader{
		app: app,
	}
}

func (s *ReadeaseReader) readFromCache(ctx context.Context, url string) (string, error) {
	article, err := GetReadeaseArticleByUrl(ctx, s.app.Dao(), url)
	if err != nil {
		return "", fmt.Errorf("summaryArticle readFromCache err: %w", err)
	}
	if article != nil {
		slog.Info("success read article from cache", "article", article.Title, "url", article.Url, "content", article.Content[:100])
		return article.Content, nil
	}
	return "", nil
}

func (s *ReadeaseReader) read(ctx context.Context, url string) (string, string, error) {
	article, err := parser.New().Parse(url)
	if err != nil {
		return "", "", fmt.Errorf("summaryArticle readArticle err: %w", err)
	}
	slog.Info("success read article", "article", article.Title, "url", article.URL, "content", article.Content[:100])

	if err := UpsertReadeaseArticle(ctx, s.app.Dao(), &ReadeaseArticle{
		Url:         url,
		OriginalUrl: article.URL,
		Title:       article.Title,
		Content:     article.Content,
	}); err != nil {
		slog.Error("upsertArticle err", "err", err)
	}

	claude := claudeweb.DefaultClaudeWeb()
	cov, err := claude.CreateConversation(article.Title)
	if err != nil {
		return "", "", fmt.Errorf("summaryArticle create conversation err: %v", err)
	}

	slog.Info("success create conversation", "conversation", cov)

	t := template.Must(template.New("promptTemplate").Parse(promptTemplate))
	var prompt strings.Builder
	if err := t.Execute(&prompt, article); err != nil {
		return "", "", fmt.Errorf("summaryArticle execute template err: %v", err)
	}
	return cov.UUID, prompt.String(), nil
}

func (s *ReadeaseReader) Read(ctx context.Context, url string) (string, error) {
	content, err := s.readFromCache(ctx, url)
	if err != nil {
		slog.Error("readFromCache err", "err", err)
	} else if content != "" {
		return content, nil
	}

	covId, prompt, err := s.read(ctx, url)
	if err != nil {
		return "", err
	}
	claude := claudeweb.DefaultClaudeWeb()
	resp, err := claude.CreateChatMessage(covId, prompt)
	if err != nil {
		return "", fmt.Errorf("summaryArticle create chat message err: %v", err)
	}
	if err := UpsertReadeaseArticle(ctx, s.app.Dao(), &ReadeaseArticle{
		Url:     url,
		Summary: resp.Completion,
	}); err != nil {
		slog.Error("upsertArticle err", "err", err)
	}
	return resp.Completion, nil
}

func (s *ReadeaseReader) ReadStream(ctx context.Context, url string, respChan chan *claudeweb.ChatMessageResponse, errChan chan error) {
	content, err := s.readFromCache(ctx, url)
	if err != nil {
		slog.Error("readFromCache err", "err", err)
	} else if content != "" {
		respChan <- &claudeweb.ChatMessageResponse{
			Completion: content,
		}
		// send EOF
		errChan <- io.EOF
		return
	}
	covId, prompt, err := s.read(ctx, url)
	if err != nil {
		errChan <- err
		return
	}
	claude := claudeweb.DefaultClaudeWeb()

	fullRespChan := make(chan string)

	claude.CreateChatMessageStreamWithFullResponse(covId, prompt, respChan, fullRespChan, errChan)

	select {
	case fullResp := <-fullRespChan:
		if err := UpsertReadeaseArticle(ctx, s.app.Dao(), &ReadeaseArticle{
			Url:     url,
			Summary: fullResp,
		}); err != nil {
			slog.Error("upsertArticle err", "err", err)
		}
	}
}
