package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/template"
	"time"

	"aienvoy/internal/pkg/dao"
	"aienvoy/internal/pkg/logger"
	"aienvoy/internal/pkg/parser"
	"aienvoy/pkg/claudeweb"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	tb "gopkg.in/telebot.v3"
)

const HOST_HN = "news.ycombinator.com"

var InvalidURLError = errors.New("Inavlid url")

var promptTemplate = `
You are a college professor. You are reading an essay.
Here is the essay: \n<<essay>>\n{{.Content}}\n<<essay>>\n
lease do the following job: 1): write a short paragraph to summarize the essay. 2): Then take notes on the essay using bullet points. 3): Write down the main point of the essay and explain in details how the author prove the point. 4): Write down the main idea of each paragraph . 5): Write a Q&A notes, Ask 10 question about the essay and get answer from essay.
Please response in Chinese and do not make other extra explaination and comment. An example of response will be:
<<example response>>
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
<<example response>>
`

func OnText(c tb.Context) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 60*5*time.Second)
	defer cancel()
	_, err := url.ParseRequestURI(c.Text())
	if err != nil || !strings.HasPrefix(c.Text(), "http") {
		return c.Reply(fmt.Sprintf("Invalid url %s, please check and try again", c.Text()))
	}

	msg, err := c.Bot().Send(c.Sender(), "Please wait a moment, I am reading the article...")
	if err != nil {
		return fmt.Errorf("Summary article err: %v", err)
	}

	var record *dao.ReadeaseArticle

	app := c.Get("app").(*pocketbase.PocketBase)
	if err := app.Dao().RunInTransaction(func(tx *daos.Dao) error {
		record, err = dao.GetReadeaseArticleByUrl(tx, c.Text())
		return err
	}); err != nil {
		logger.SugaredLogger.Warnw("get article from DB error", "err", err)
	}

	if record != nil {
		logger.SugaredLogger.Infow("get article from DB", "article", record)
		return c.Send(record.Summary)
	}
	respChan := make(chan *claudeweb.ChatMessageResponse)
	errChan := make(chan error)
	defer close(respChan)
	defer close(errChan)

	go summaryArticle(ctx, c.Text(), respChan, errChan)

	text := ""
	chunk := ""

	for {
		select {
		case resp := <-respChan:
			text += resp.Completion
			chunk += resp.Completion
			if strings.TrimSpace(chunk) == "" {
				continue
			}
			if len(chunk) > 200 {
				// logger.SugaredLogger.Debugw("response with text", "text", text)
				newMsg, err := c.Bot().Edit(msg, text)
				if err != nil {
					logger.SugaredLogger.Warnw("OnText edit msg err", "err", err)
				} else {
					msg = newMsg
				}
				chunk = ""
			}
		case err := <-errChan:
			if errors.Is(err, InvalidURLError) {
				return c.Reply("Invalid url, please check and try again")
			} else if errors.Is(err, io.EOF) {

				// send last message
				if _, err := c.Bot().Edit(msg, text); err != nil {
					logger.SugaredLogger.Errorw("OnText edit msg err", "err", err)
					return err
				}
				// save record to db
				if err := app.Dao().RunInTransaction(func(tx *daos.Dao) error {
					return dao.CreateReadeaseArticle(tx, &dao.ReadeaseArticle{
						Url:     c.Text(),
						Summary: text,
					})
				}); err != nil {
					logger.SugaredLogger.Warnw("save article to DB error", "err", err)
				}
				return nil
			}
			if _, err = c.Bot().Edit(msg, text); err != nil {
				logger.SugaredLogger.Errorw("OnText edit msg err", "err", err)
			}
			return fmt.Errorf("Summary article err: %v", err)
		case <-ctx.Done():
			logger.SugaredLogger.Errorw("OnText timeout", "err", ctx.Err())
			return fmt.Errorf("Summary article timeout, please wait a moment and try again")
		}
	}
}

func summaryArticle(ctx context.Context, uri string, respChan chan *claudeweb.ChatMessageResponse, errChan chan error) {
	article, err := parser.New().Parse(uri)
	if err != nil {
		logger.SugaredLogger.Errorw("read article err", "err", err)
		errChan <- fmt.Errorf("summaryArticle readArticle err: %w", err)
		return
	}
	logger.SugaredLogger.Infow("success read article", "article", article.Title, "url", article.URL, "content", article.Content[:100])

	claude := claudeweb.DefaultClaudeWeb()
	cov, err := claude.CreateConversation(article.Title)
	if err != nil {
		errChan <- fmt.Errorf("summaryArticle create conversation err: %v", err)
		return
	}

	logger.SugaredLogger.Infow("success create conversation", "conversation", cov)

	t := template.Must(template.New("promptTemplate").Parse(promptTemplate))
	var prompt strings.Builder
	if err := t.Execute(&prompt, article); err != nil {
		errChan <- fmt.Errorf("summaryArticle execute template err: %v", err)
		return
	}

	claude.CreateChatMessageStream(cov.UUID, prompt.String(), respChan, errChan)
}
