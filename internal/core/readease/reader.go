package readease

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"strings"

	innerllm "github.com/Vaayne/aienvoy/internal/core/llm"
	"github.com/Vaayne/aienvoy/internal/pkg/parser"
	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/pocketbase/pocketbase"
)

var prompt = `
----
Please complete the following tasks and provide responses in Chinese enclosed within corresponding XML tags:
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

type Reader struct {
	app *pocketbase.PocketBase
}

func NewReader(app *pocketbase.PocketBase) *Reader {
	return &Reader{
		app: app,
	}
}

func (s *Reader) read(ctx context.Context, url string) (*Article, error) {
	var content parser.Content
	var err error

	article, err := GetArticleByUrl(ctx, s.app.Dao(), url)
	if err != nil {
		slog.ErrorContext(ctx, "get article from db error", "err", err)
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
			slog.WarnContext(ctx, "did not get any content from the url", "url", url, "title", content.Title)
			return nil, fmt.Errorf("did not get any content from url %s", url)
		}

		if content.URL == "" {
			content.URL = url
		}

		article = &Article{
			Url:         url,
			OriginalUrl: content.URL,
			Title:       content.Title,
			Content:     content.Content,
		}

		if err := UpsertArticle(ctx, s.app.Dao(), article); err != nil {
			slog.ErrorContext(ctx, "upsertArticle err", "err", err)
		}
	}

	slog.InfoContext(ctx, "success read article", "article", content.Title, "url", url, "content", content.Content[:min(100, len(content.Content))])
	return article, nil
}

func (s *Reader) Read(ctx context.Context, url, model string) (*Article, error) {
	// get article content by parse web page
	article, err := s.read(ctx, url)
	if err != nil {
		return nil, err
	}

	if article != nil && article.Summary != "" {
		slog.DebugContext(ctx, "article alreay summaried", "url", url, "title", article.Title, "summary", article.Summary[:100])
		return article, nil
	}

	// summary article
	llmSvc := innerllm.New(model, innerllm.NewDao(s.app.Dao()))
	if llmSvc == nil {
		slog.Error("failed to create llm service", "model", model)
		return nil, fmt.Errorf("failed to create llm service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Model:       model,
		Messages:    buildMessages(article),
		MaxTokens:   8192,
		Temperature: 0.7,
	}
	resp, err := llmSvc.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("summaryArticle create chat message err: %v", err)
	}

	summary, err := buildSummaryResponse(url, article.Title, resp.Choices[0].Message.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to build summary response %w", err)
	}

	article.Summary = summary
	article.LlmModel = model

	// save result to cache
	if err := UpsertArticle(ctx, s.app.Dao(), article); err != nil {
		slog.Error("upsertArticle err", "err", err)
	}
	return article, nil
}

func (s *Reader) ReadStream(ctx context.Context, url, model string, respChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	article, err := s.read(ctx, url)
	if err != nil {
		errChan <- err
		return
	}

	if article != nil && article.Summary != "" {
		slog.InfoContext(ctx, "article already summaries", "url", url, "title", article.Title, "summary", article.Summary[:100])
		respChan <- llm.ChatCompletionStreamResponse{}
		return
	}

	llmSvc := innerllm.New(model, innerllm.NewDao(s.app.Dao()))
	if llmSvc == nil {
		slog.ErrorContext(ctx, "failed to create llm service", "model", model)
		errChan <- fmt.Errorf("failed to create llm service: %w", err)
		return
	}

	req := llm.ChatCompletionRequest{
		Model:       model,
		Messages:    buildMessages(article),
		MaxTokens:   8192,
		Temperature: 0.7,
		Stream:      true,
	}

	dataChan := make(chan llm.ChatCompletionStreamResponse)
	// defer close(dataChan)
	innerErrChan := make(chan error)
	// defer close(innerErrChan)

	go llmSvc.CreateChatCompletionStream(ctx, req, dataChan, innerErrChan)
	sb := strings.Builder{}
	for {
		select {
		case resp := <-dataChan:
			sb.WriteString(resp.Choices[0].Delta.Content)
			respChan <- resp
		case err := <-innerErrChan:
			if errors.Is(err, io.EOF) {
				summary, err := buildSummaryResponse(url, article.Title, sb.String())
				if err != nil {
					errChan <- fmt.Errorf("failed to build summary response %w", err)
					return
				}
				article.Summary = summary
				article.LlmModel = req.Model

				if err := UpsertArticle(ctx, s.app.Dao(), article); err != nil {
					slog.ErrorContext(ctx, "upsertArticle err", "err", err)
				}
				slog.InfoContext(ctx, "success stream summary article", "url", url, "title", article.Title, "summary", article.Summary[:100])
			}
			errChan <- err
			return
		case <-ctx.Done():
			slog.WarnContext(ctx, "readease reader stream context done")
		}
	}
}

func buildMessages(article *Article) []llm.ChatCompletionMessage {
	return []llm.ChatCompletionMessage{
		{
			Role:    llm.ChatMessageRoleSystem,
			Content: "You are my reading partner, you will help me summarize the article enclosed within the XML tag <article>.",
		},
		{
			Role:    llm.ChatMessageRoleUser,
			Content: fmt.Sprintf("Here is my article enclosed within the XML tag <article>.\n<article>%s</article>", article.Content),
		},
		{
			Role:    llm.ChatMessageRoleUser,
			Content: prompt,
		},
	}
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
