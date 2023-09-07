package bard

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	bardUrl        = "https://bard.google.com/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate"
	cookieTokenKey = "__Secure-1PSID"
)

var headers map[string]string = map[string]string{
	"Host":          "bard.google.com",
	"X-Same-Domain": "1",
	"User-Agent":    "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Content-Type":  "application/x-www-form-urlencoded;charset=utf-8",
	"Origin":        "https://bard.google.com",
	"Referer":       "https://bard.google.com/",
}

type BardClient struct {
	mu      sync.Mutex
	token   string
	timeout time.Duration
	session *http.Client
}

type BardClientOption func(*BardClient)

func NewBardClient(token string, opts ...BardClientOption) (*BardClient, error) {
	if token == "" || !strings.HasSuffix(token, ".") {
		return nil, fmt.Errorf("__Secure-1PSID value must end with a single dot. Enter correct __Secure-1PSID value.")
	}

	b := &BardClient{
		token:   token,
		timeout: 10 * time.Second,
		session: &http.Client{},
	}

	for _, opt := range opts {
		opt(b)
	}

	if b.timeout != 0 {
		b.session.Timeout = b.timeout
	}

	if _, err := b.getSNlM0e(); err != nil {
		return nil, fmt.Errorf("init bard error %w", err)
	}

	return b, nil
}

func WithTimeout(timeout time.Duration) BardClientOption {
	return func(bc *BardClient) {
		bc.timeout = timeout
	}
}

func WithSession(session *http.Client) BardClientOption {
	return func(bc *BardClient) {
		bc.session = session
	}
}

func (b *BardClient) getSNlM0e() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	req, err := http.NewRequest(http.MethodGet, "https://bard.google.com/", nil)
	if err != nil {
		return "", fmt.Errorf("init request for SNlM0e error: %w", err)
	}
	b.setHeaders(req)

	resp, err := b.session.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Response code not 200. Response Status is %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`SNlM0e":"(.*?)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		slog.Warn("SNlM0e value not found in response.", "resp", string(body))
		return "", fmt.Errorf("SNlM0e value not found in response. Check __Secure-1PSID value.")
	}
	return matches[1], nil
}

func (b *BardClient) Ask(prompt, conversationID, responseID, choiceID string, reqID int) (*BardAnswer, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	req, err := b.buildRequest(prompt, conversationID, responseID, choiceID, reqID)
	if err != nil {
		return nil, fmt.Errorf("build bard request error: %w", err)
	}
	resp, err := b.session.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to bard error: %w", err)
	}
	defer resp.Body.Close()
	return b.parseResponse(req.Body)
}

func (b *BardClient) setHeaders(req *http.Request) {
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	req.Header.Set("Cookie", fmt.Sprintf("%s=%s", cookieTokenKey, b.token))
}

func (b *BardClient) buildRequest(prompt, conversationID, responseID, choiceID string, reqID int) (*http.Request, error) {
	// build req url
	if reqID == 0 {
		reqID = rand.Intn(10000)
	}
	params := url.Values{
		"bl":     {"boq_assistant-bard-web-server_20230510.09_p1"},
		"_reqid": {strconv.Itoa(reqID)},
		"rt":     {"c"},
	}
	reqURL := bardUrl + "?" + params.Encode()

	// build req body
	// get snlm0e
	snlm0e, err := b.getSNlM0e()
	if err != nil {
		return nil, err
	}

	inputTextStruct := [][]any{
		{prompt},
		nil,
		{conversationID, responseID, choiceID},
	}

	inputText, err := json.Marshal(inputTextStruct)
	if err != nil {
		return nil, fmt.Errorf("encode input text error: %w", err)
	}

	values := &url.Values{}
	values.Add("f.req", fmt.Sprintf("[null,%s]", inputText))
	values.Add("at", snlm0e)
	reqBody := strings.NewReader(values.Encode())

	// new http request
	req, err := http.NewRequest(http.MethodPost, reqURL, reqBody)
	if err != nil {
		return nil, err
	}
	// set header
	b.setHeaders(req)

	return req, nil
}

func (b *BardClient) parseResponse(r io.ReadCloser) (*BardAnswer, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read bard response body error: %w", err)
	}
	var respDict []any
	err = json.Unmarshal(body, &respDict)
	if err != nil {
		return nil, fmt.Errorf("unmarshal bard response body error: %w", err)
	}
	if len(respDict) < 4 || respDict[3] == nil {
		return nil, fmt.Errorf("bard response error: %s", string(body))
	}

	parsedAnswer := respDict[3].([]any)
	answer := &BardAnswer{
		Content:        parsedAnswer[0].(string),
		ConversationID: parsedAnswer[1].([]any)[0].(string),
		ResponseID:     parsedAnswer[1].([]any)[1].(string),
	}
	if parsedAnswer[3] != nil {
		answer.FactualityQueries = parsedAnswer[3].([]interface{})
	}

	if parsedAnswer[2] != nil {
		answer.TextQuery = parsedAnswer[2].([]interface{})[0].(string)
	}

	if parsedAnswer[4] != nil {
		choices := parsedAnswer[4].([]interface{})
		answer.Choices = make([]Choice, len(choices))
		for i, choice := range choices {
			choiceMap := choice.([]interface{})
			answer.Choices[i] = Choice{
				ID:      choiceMap[0].(string),
				Content: choiceMap[1].(string),
			}
		}
	}
	return answer, nil
}
