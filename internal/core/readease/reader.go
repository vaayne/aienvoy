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
As a college professor, you're tasked with analyzing an essay. Here is the essay content:
<essay>
{{.Content}}
</essay>
Please execute the following tasks:
1. Create a bullet point outline summarizing the essay.
2. Identify and elaborate on the main point of the essay, explaining how the author substantiates it.
3. Summarize each paragraph.
4. Develop a Q&A note with 10 questions about the essay, and provide answers derived from the essay.

An example of response will be:
<ResponseExample>
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
<Q&A>
- Q: What is the main point of this essay?
- A: The main point of this essay is how to write a summary.
- Q: What is the main idea of the first paragraph?
- A: The main idea of the first paragraph is ...
</Q&A>
</ResponseExample>

Please respond in Chinese and refrain from including any additional explanations or comments and keep these xml tags in English as they were: <Summary>, <MainPoint>, <ParagraphSummary>, <Q&A>
`

type ReadeaseReader struct {
	app *pocketbase.PocketBase
}

func NewReader(app *pocketbase.PocketBase) *ReadeaseReader {
	return &ReadeaseReader{
		app: app,
	}
}

func (s *ReadeaseReader) readSummaryFromCache(ctx context.Context, url string) string {
	article, err := GetReadeaseArticleByUrl(ctx, s.app.Dao(), url)
	if err == nil && article != nil {
		return article.Summary
	}
	slog.Warn("Failed read article from cache", "url", url)
	return ""
}

func (s *ReadeaseReader) read(ctx context.Context, url string) (string, string, error) {
	var article parser.Content
	var err error

	record, err := GetReadeaseArticleByUrl(ctx, s.app.Dao(), url)
	if err == nil && record != nil {
		article = parser.Content{
			URL:     record.OriginalUrl,
			Title:   record.Title,
			Content: record.Content,
		}
	}

	if article.Content == "" {
		article, err = parser.New().Parse(url)
		if err != nil {
			return "", "", fmt.Errorf("summaryArticle readArticle err: %w", err)
		}
		if err := UpsertReadeaseArticle(ctx, s.app.Dao(), &ReadeaseArticle{
			Url:         url,
			OriginalUrl: article.URL,
			Title:       article.Title,
			Content:     article.Content,
		}); err != nil {
			slog.Error("upsertArticle err", "err", err)
		}
	}
	slog.Info("success read article", "article", article.Title, "url", url, "content", article.Content[:min(100, len(article.Content))])

	claude := claudeweb.DefaultClaudeWeb()
	cov, err := claude.CreateConversation(article.Title)
	if err != nil {
		return "", "", fmt.Errorf("failed create conversation err: %v", err)
	}

	slog.Debug("success create conversation", "conversation", cov)

	t := template.Must(template.New("promptTemplate").Parse(promptTemplate))
	var prompt strings.Builder
	if err := t.Execute(&prompt, article); err != nil {
		return "", "", fmt.Errorf("summaryArticle execute template err: %v", err)
	}
	return cov.UUID, prompt.String(), nil
}

func (s *ReadeaseReader) Read(ctx context.Context, url string) (string, error) {
	// get summary from Cache
	summary := s.readSummaryFromCache(ctx, url)
	if summary != "" {
		return summary, nil
	}

	// get article content by parse web page
	covId, prompt, err := s.read(ctx, url)
	if err != nil {
		return "", err
	}
	// summary article
	claude := claudeweb.DefaultClaudeWeb()
	resp, err := claude.CreateChatMessage(covId, prompt)
	if err != nil {
		return "", fmt.Errorf("summaryArticle create chat message err: %v", err)
	}

	// save result to cache
	if err := UpsertReadeaseArticle(ctx, s.app.Dao(), &ReadeaseArticle{
		Url:      url,
		Summary:  resp.Completion,
		LlmType:  "claude",
		LlmCovId: covId,
	}); err != nil {
		slog.Error("upsertArticle err", "err", err)
	}
	return resp.Completion, nil
}

func (s *ReadeaseReader) ReadStream(ctx context.Context, url string, respChan chan *claudeweb.ChatMessageResponse, errChan chan error) {
	summary := s.readSummaryFromCache(ctx, url)
	if summary != "" {
		respChan <- &claudeweb.ChatMessageResponse{
			Completion: summary,
		}
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

	if err := UpsertReadeaseArticle(ctx, s.app.Dao(), &ReadeaseArticle{
		Url:      url,
		Summary:  <-fullRespChan,
		LlmType:  "claude",
		LlmCovId: covId,
	}); err != nil {
		slog.Error("upsertArticle err", "err", err)
	}
}
