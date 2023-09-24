package parser

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/Vaayne/aienvoy/pkg/hackernews"
)

type HackerNewsParser struct{}

func (p HackerNewsParser) Parse(uri string) (Content, error) {
	var content Content
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return content, InvalidURLError
	}
	id := u.Query().Get("id")
	if id == "" {
		return content, fmt.Errorf("Invalid hacker news url %s", u.String())
	}
	itemId, err := strconv.Atoi(id)
	if err != nil {
		return content, fmt.Errorf("Invalid hacker news url %s", u.String())
	}
	hn := hackernews.New()
	item, err := hn.GetItem(itemId)
	if err != nil {
		return content, fmt.Errorf("Error getting item %d, error: %w", itemId, err)
	}

	if item.Url != "" && item.Text == "" {
		// If the item has a url, but no text, we need to parse the original content
		parser := DefaultParser{}
		originalContent, err := parser.Parse(item.Url)
		if err != nil {
			return content, fmt.Errorf("Error parsing url %s, error: %w", item.Url, err)
		}
		item.Text = originalContent.Content
	}

	return hackerNewsItemToContent(item), nil
}

func hackerNewsItemToContent(item hackernews.Item) Content {
	return Content{
		URL:     item.Url,
		Title:   item.Title,
		Content: item.Text,
	}
}
