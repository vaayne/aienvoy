package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var InvalidURLError = errors.New("Inavlid url")

type DefaultParser struct{}

func New() *DefaultParser {
	return &DefaultParser{}
}

const BaseUrl = "https://readability.theboys.tech/api/parser?url="

const HackerNewsHost = "news.ycombinator.com"

func (p *DefaultParser) Parse(urlStr string) (Content, error) {
	var content Content

	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return content, InvalidURLError
	}

	switch u.Host {
	case HackerNewsHost:
		parser := HackerNewsParser{}
		return parser.Parse(urlStr)
	}

	uri := BaseUrl + urlStr
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return content, fmt.Errorf("Error creating request for url: %s, error: %w", uri, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return content, fmt.Errorf("Error making request for url: %s, error: %w", uri, err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return content, fmt.Errorf("Error making request for url: %s, status code: %d", uri, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return content, fmt.Errorf("Error reading response for url: %s, error: %w", uri, err)
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, &content)
	if err != nil {
		return content, fmt.Errorf("Error decoding response for url: %s, error: %w", uri, err)
	}
	return content, nil
}
