package readease

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm/llmclaudeweb"
	"github.com/Vaayne/aienvoy/internal/pkg/parser"
	"github.com/Vaayne/aienvoy/pkg/claudeweb"

	"github.com/pocketbase/pocketbase"
)

var promptTemplate = `
You are my reading partner, and I have an article enclosed within the XML tag <article>:

<article>
{{.Content}}
</article>

----
Please complete the following tasks and provide responses enclosed within corresponding XML tags:
1. Summarize each paragraph in the article separately. Use <ParagraphSummary> ... </ParagraphSummary> for all paragraph summaries, and label each paragraph with its number.
2. Summarize this article as a whole. Use <Summary> ... </Summary> for the overall article summary.
3. Create a list of study questions and answers. Use <StudyQuestions> ... </StudyQuestions> for the list, and format each question and answer using Q: and A: labels.
----
For reference, here's an example response format:
<ParagraphSummary>
- Paragraph 1: ...
- Paragraph 2: ...
...
</ParagraphSummary>

<Summary>
...
</Summary>

<StudyQuestions>
Q: ...
A: ...
Q: ...
A: ...
...
</StudyQuestions>
----
Please ensure that you maintain the XML tags in your responses and avoid adding explanations or extra words.
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
	if err != nil {
		slog.Error("get article from db error", "err", err)
	}
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

		if content.URL == "" {
			content.URL = url
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
	claude := llmclaudeweb.New()
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

	claude := llmclaudeweb.New()
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
