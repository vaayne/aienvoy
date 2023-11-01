package phind

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Vaayne/aienvoy/pkg/session"
	utls "github.com/refraction-networking/utls"
	"github.com/sashabaranov/go-openai"
)

const (
	host      = "https://www.phind.com"
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"
)

var clientHelloID = utls.HelloChrome_100_PSK

type Phind struct {
	session *session.Session
	cookies []*http.Cookie
}

func New(cookies []*http.Cookie) *Phind {
	return &Phind{
		session: session.New(session.WithClientHelloID(clientHelloID)),
		cookies: cookies,
	}
}

func (p *Phind) createCompletion(ctx context.Context, payload *Request, respChan chan openai.ChatCompletionStreamResponse, errChan chan error) {
	uri := host + "/api/agent"
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		errChan <- fmt.Errorf("createCompletion marshal payload err: %w", err)
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(payloadBytes))
	if err != nil {
		errChan <- fmt.Errorf("createCompletion new request err: %w", err)
		return
	}
	resp, err := p.request(req)
	if err != nil {
		errChan <- fmt.Errorf("createCompletion do request err: %w", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		var msg any
		_ = json.NewDecoder(resp.Body).Decode(&msg)
		slog.Error("createCompletion response status code err", "status", resp.StatusCode, "msg", msg)
		errChan <- fmt.Errorf("createCompletion response status code err: %d", resp.StatusCode)
		return
	}
	var data openai.ChatCompletionStreamResponse
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				errChan <- err
				return
			}
			errChan <- fmt.Errorf("createCompletion read response body err: %w", err)
			return
		}
		if len(line) > 6 {
			err = json.Unmarshal(line[6:], &data)

			if err != nil {
				slog.Error("json unmarshal error", "err", err)
				errChan <- fmt.Errorf("createCompletion unmarshal response body err: %w", err)
				return
			}
			slog.Info("createCompletion response", "data", data)
			respChan <- data
		}
	}
}

func (p *Phind) request(req *http.Request) (*http.Response, error) {
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
