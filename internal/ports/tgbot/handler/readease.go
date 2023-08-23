package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"aienvoy/internal/pkg/dao"
	"aienvoy/internal/pkg/logger"
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
	article, err := readArticle(ctx, uri)
	if err != nil {
		logger.SugaredLogger.Errorw("read article err", "err", err)
		errChan <- fmt.Errorf("summaryArticle readArticle err: %w", err)
		return
	}
	logger.SugaredLogger.Infow("success read article", "article", article.Title, "url", article.Url, "content", article.Content[:100])

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

type Article struct {
	Url     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func readArticle(ctx context.Context, uri string) (*Article, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, InvalidURLError
	}
	logger.SugaredLogger.Infow("success parse url", "url", u.String())
	article := &Article{
		Url: u.String(),
	}
	switch u.Host {
	case HOST_HN:
		hn, err := readHackerNews(ctx, u)
		if err != nil {
			return nil, fmt.Errorf("readHackerNews err: %v", err)
		}
		article.Title = hn.Title
		article.Content = hn.Content
		if article.Content == "" {
			article.Content = hn.OriginalContent
		}
	default:
		content, err := parseUrlContent(ctx, uri)
		if err != nil {
			return nil, fmt.Errorf("parseUrlContent err: %v", err)
		}
		article.Title = content.Title
		article.Content = content.Content
	}
	return article, nil
}

type HackerNews struct {
	Id              int    `json:"id"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	OriginalUrl     string `json:"original_url"`
	OriginalTitle   string `json:"original_title"`
	OriginalContent string `json:"original_content"`
}

func readHackerNews(ctx context.Context, u *url.URL) (*HackerNews, error) {
	id := u.Query().Get("id")
	if id == "" {
		return nil, fmt.Errorf("Invalid hacker news url %s", u.String())
	}

	meta, err := readHackerNewsMeta(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("readHackerNewsMeta err: %v", err)
	}

	hn := &HackerNews{
		Id:          meta.Id,
		Title:       meta.Title,
		Content:     meta.Text,
		OriginalUrl: meta.Url,
	}

	logger.SugaredLogger.Infow("success read hacker news meta", "meta", meta)
	if hn.OriginalUrl != "" && hn.Content == "" {
		content, err := parseUrlContent(ctx, hn.OriginalUrl)
		if err != nil {
			return nil, fmt.Errorf("parseUrlContent err: %v", err)
		}
		hn.OriginalTitle = content.Title
		hn.OriginalContent = content.Content
	}
	return hn, nil
}

// HackerNewsMeta is the meta data of a hacker news item
// https://github.com/HackerNews/API#items
/*
	id	The item's unique id.
	deleted	true if the item is deleted.
	type	The type of item. One of "job", "story", "comment", "poll", or "pollopt".
	by	The username of the item's author.
	time	Creation date of the item, in Unix Time.
	text	The comment, story or poll text. HTML.
	dead	true if the item is dead.
	parent	The comment's parent: either another comment or the relevant story.
	poll	The pollopt's associated poll.
	kids	The ids of the item's comments, in ranked display order.
	url	The URL of the story.
	score	The story's score, or the votes for a pollopt.
	title	The title of the story, poll or job. HTML.
	parts	A list of related pollopts, in display order.
	descendants	In the case of stories or polls, the total comment count.
*/
type HackerNewsMeta struct {
	Id          int      `json:"id"`
	Deleted     bool     `json:"deleted"`
	Type        string   `json:"type"`
	By          string   `json:"by"`
	Time        int      `json:"time"`
	Text        string   `json:"text"`
	Dead        bool     `json:"dead"`
	Parent      string   `json:"parent"`
	Poll        string   `json:"poll"`
	Kids        []int    `json:"kids"`
	Url         string   `json:"url"`
	Score       int      `json:"score"`
	Title       string   `json:"title"`
	Parts       []string `json:"parts"`
	Descendants int      `json:"descendants"`
}

func readHackerNewsMeta(ctx context.Context, id string) (*HackerNewsMeta, error) {
	uri := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%s.json", id)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("readHackerNewsMeta NewRequest err: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("readHackerNewsMeta error send http request GET %s, err: %v", uri, err)
	}
	defer resp.Body.Close()

	var meta HackerNewsMeta

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("readHackerNewsMeta read response bytes err: %v", err)
	}
	err = json.Unmarshal(body, &meta)
	if err != nil {
		logger.SugaredLogger.Errorw("readHackerNewsMeta Unmarshal err", "err", err, "body", string(body))
		return nil, fmt.Errorf("readHackerNewsMeta Unmarshal err: %v", err)
	}

	return &meta, nil
}

// ParseUrlContent is the content of a url
// https://github.com/postlight/parser
/*
{
  "title": "Thunder (mascot)",
  "content": "... <p><b>Thunder</b> is the <a href=\"https://en.wikipedia.org/wiki/Stage_name\">stage name</a> for the...",
  "author": "Wikipedia Contributors",
  "date_published": "2016-09-16T20:56:00.000Z",
  "lead_image_url": null,
  "dek": null,
  "next_page_url": null,
  "url": "https://en.wikipedia.org/wiki/Thunder_(mascot)",
  "domain": "en.wikipedia.org",
  "excerpt": "Thunder Thunder is the stage name for the horse who is the official live animal mascot for the Denver Broncos",
  "word_count": 4677,
  "direction": "ltr",
  "total_pages": 1,
  "rendered_pages": 1
}
*/
type ParseUrlContent struct {
	Title         string `json:"title"`
	Content       string `json:"content"`
	Author        string `json:"author"`
	DatePublished string `json:"date_published"`
	LeadImageURL  string `json:"lead_image_url"`
	Dek           string `json:"dek"`
	NextPageURL   string `json:"next_page_url"`
	URL           string `json:"url"`
	Domain        string `json:"domain"`
	Excerpt       string `json:"excerpt"`
	WordCount     int    `json:"word_count"`
	Direction     string `json:"direction"`
	TotalPages    int    `json:"total_pages"`
	RenderedPages int    `json:"rendered_pages"`
}

func parseUrlContent(ctx context.Context, uri string) (*ParseUrlContent, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://readability.theboys.tech/api/parser?url=%s", uri), nil)
	if err != nil {
		return nil, fmt.Errorf("parseUrlContent NewRequest err: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("parseUrlContent error send http request GET %s, err: %v", uri, err)
	}
	defer resp.Body.Close()
	var content ParseUrlContent
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parseUrlContent read response bytes err: %v", err)
	}
	err = json.Unmarshal(body, &content)
	if err != nil {
		return nil, fmt.Errorf("parseUrlContent Unmarshal err: %v", err)
	}
	return &content, nil
}
