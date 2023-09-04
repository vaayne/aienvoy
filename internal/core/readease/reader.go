package readease

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"strings"

	"aienvoy/internal/pkg/parser"
	"aienvoy/pkg/claudeweb"

	"github.com/pocketbase/pocketbase"
)

var promptTemplate = `
As a college professor, you're tasked with analyzing an essay. Here is the essay content:
<essay>
{{.Content}}
</essay>
Please execute the following tasks:
1. Create a bullet point outline summarizing the essay.
2. Identify and elaborate on the main point of the essay, explaining how the author substantiates it.
3. Summarize each paragraph.

An example of response will be:
<Summary>
This essay is about how to write a summary
- The author first introduce the importance of summary and then explain how to write a summary.
- ...
</Summary>
<MainPoint>
The main point of this essay is how to write a summary.
The author prove the point by ...
</MainPoint>
<ParagraphSummary>
- The first paragraph is saying ...
- The second paragraph is about ...
- ...
</ParagraphSummary>

Please respond in Chinese and refrain from including any additional explanations or comments and keep these xml tags in English as they were: <Summary>, <MainPoint>, <ParagraphSummary>
`

type ReadeaseReader struct {
	app *pocketbase.PocketBase
}

func NewReader(app *pocketbase.PocketBase) *ReadeaseReader {
	return &ReadeaseReader{
		app: app,
	}
}

func (s *ReadeaseReader) read(ctx context.Context, url string) (*ReadeaseArticle, error) {
	var content parser.Content
	var err error

	article, err := GetReadeaseArticleByUrl(ctx, s.app.Dao(), url)
	if err == nil && article != nil {
		content = parser.Content{
			URL:     article.OriginalUrl,
			Title:   article.Title,
			Content: article.Content,
		}
	}

	if content.Content == "" {
		content, err = parser.New().Parse(url)
		if err != nil {
			return nil, fmt.Errorf("summaryArticle readArticle err: %w", err)
		}

		if content.Content == "" {
			slog.Warn("did not get any content from the url", "url", url, "title", content.Title)
			return nil, fmt.Errorf("did not get any content from url %s", url)
		}

		article = &ReadeaseArticle{
			Url:         url,
			OriginalUrl: content.URL,
			Title:       content.Title,
			Content:     content.Content,
		}

		if err := UpsertReadeaseArticle(ctx, s.app.Dao(), article); err != nil {
			slog.Error("upsertArticle err", "err", err)
		}
	}

	slog.Info("success read article", "article", content.Title, "url", url, "content", content.Content[:min(100, len(content.Content))])
	return article, nil
}

func (s *ReadeaseReader) Read(ctx context.Context, url string) (*ReadeaseArticle, error) {
	// get article content by parse web page
	article, err := s.read(ctx, url)
	if err != nil {
		return nil, err
	}

	if article != nil && article.Summary != "" {
		slog.Debug("article alreay summaried", "url", url, "title", article.Title, "summary", article.Summary[:100])
		return article, nil
	}

	// summary article
	claude := claudeweb.DefaultClaudeWeb()
	cov, err := claude.CreateConversation(article.Title)
	if err != nil {
		return nil, fmt.Errorf("failed create conversation err: %v", err)
	}
	slog.Debug("success create conversation", "conversation", cov)

	prompt, err := buildPrompt(article)
	if err != nil {
		return nil, fmt.Errorf("failed build prompt for summary, %w", err)
	}

	resp, err := claude.CreateChatMessage(cov.UUID, prompt)
	if err != nil {
		return nil, fmt.Errorf("summaryArticle create chat message err: %v", err)
	}

	summary, err := buildSummaryResponse(url, article.Title, resp.Completion)
	if err != nil {
		return nil, fmt.Errorf("failed to build summary response %w", err)
	}

	article.Summary = summary
	article.LlmCovId = cov.UUID
	article.LlmModel = "claude"

	// save result to cache
	if err := UpsertReadeaseArticle(ctx, s.app.Dao(), article); err != nil {
		slog.Error("upsertArticle err", "err", err)
	}
	return article, nil
}

func (s *ReadeaseReader) ReadStream(ctx context.Context, url string, respChan chan *claudeweb.ChatMessageResponse, errChan chan error) {
	article, err := s.read(ctx, url)
	if err != nil {
		errChan <- err
		return
	}

	if article != nil && article.Summary != "" {
		slog.Debug("article alreay summaried", "url", url, "title", article.Title, "summary", article.Summary[:100])
		respChan <- &claudeweb.ChatMessageResponse{
			Completion: article.Summary,
		}
		return
	}

	claude := claudeweb.DefaultClaudeWeb()
	cov, err := claude.CreateConversation(article.Title)
	if err != nil {
		errChan <- fmt.Errorf("failed create conversation err: %w", err)
		return
	}
	prompt, err := buildPrompt(article)
	if err != nil {
		errChan <- fmt.Errorf("failed build prompt for summary, %w", err)
		return
	}

	fullRespChan := make(chan string)
	defer close(fullRespChan)
	go claude.CreateChatMessageStreamWithFullResponse(cov.UUID, prompt, respChan, fullRespChan, errChan)

	select {
	case resp := <-fullRespChan:
		// save summary from AI
		summary, err := buildSummaryResponse(url, article.Title, resp)
		if err != nil {
			errChan <- fmt.Errorf("failed to build summary response %w", err)
			return
		}

		article.Summary = summary
		article.LlmCovId = cov.UUID
		article.LlmModel = "claude"

		if err := UpsertReadeaseArticle(ctx, s.app.Dao(), article); err != nil {
			slog.Error("upsertArticle err", "err", err)
		}
	case <-ctx.Done():
		slog.Warn("context done")
	}
}

func buildPrompt(article *ReadeaseArticle) (string, error) {
	t := template.Must(template.New("promptTemplate").Parse(promptTemplate))
	var prompt strings.Builder
	if err := t.Execute(&prompt, article); err != nil {
		return "", fmt.Errorf("summaryArticle execute template err: %v", err)
	}
	return prompt.String(), nil
}

func buildSummaryResponse(url, title, summary string) (string, error) {
	var sb strings.Builder
	const summaryTemplate = `
{{.Title}} ({{.Url}})

{{.Summary}}
`
	t := template.Must(template.New("summaryTemplate").Parse(summaryTemplate))

	type Summary struct {
		Url     string
		Title   string
		Summary string
	}

	err := t.Execute(&sb, &Summary{
		Url:     url,
		Title:   title,
		Summary: summary,
	})
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}
