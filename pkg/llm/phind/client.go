package phind

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/Vaayne/aienvoy/pkg/session"
	utls "github.com/refraction-networking/utls"
)

const (
	host      = "https://www.phind.com"
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"
)

var clientHelloID = utls.HelloChrome_100_PSK

type Client struct {
	session *session.Session
	cookies []*http.Cookie
}

func NewClient(cookies []*http.Cookie) *Client {
	return &Client{
		session: session.New(session.WithClientHelloID(clientHelloID)),
		cookies: cookies,
	}
}

func (p *Client) CreateCompletion(ctx context.Context, payload *Request, respChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	uri := host + "/api/agent"
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		errChan <- fmt.Errorf("phind create completion marshal payload err: %w", err)
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(payloadBytes))
	if err != nil {
		errChan <- fmt.Errorf("phind create completion new request err: %w", err)
		return
	}
	resp, err := p.request(req)
	if err != nil {
		errChan <- fmt.Errorf("phind create completion do request err: %w", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		var msg any
		err = json.NewDecoder(resp.Body).Decode(&msg)
		if err != nil {
			errChan <- fmt.Errorf("phind create completion decode response body err: %w", err)
			return
		}
		errChan <- fmt.Errorf("phind create completion response error, status code: %d", resp.StatusCode)
		return
	}
	var data llm.ChatCompletionStreamResponse
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				errChan <- err
				return
			}
			errChan <- fmt.Errorf("phind create completion read response body err: %w", err)
			return
		}
		if len(line) > 6 {
			err = json.Unmarshal(line[6:], &data)

			if err != nil {
				errChan <- fmt.Errorf("phind create completion unmarshal response body err: %w", err)
				return
			}
			respChan <- data
		}
	}
}

func (p *Client) request(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("origin", host)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Referer", host+"/")
	req.Header.Add("User-Agent", userAgent)

	for _, ck := range p.cookies {
		req.AddCookie(ck)
	}
	return p.session.Do(req)
}
