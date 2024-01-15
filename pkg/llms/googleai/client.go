package googleai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

const defaultHost = "https://generativelanguage.googleapis.com"

type Client struct {
	sess   *http.Client
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{
		sess:   http.DefaultClient,
		apiKey: apiKey,
	}
}

func (c *Client) ListModels() []string {
	return []string{"gemini-pro", "gemini-pro-vision"}
}

func (c *Client) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	reqBody := ChatRequest{}.FromChatCompletionRequest(req)
	chatResp, err := c.post(req.Model, reqBody, false)
	if err != nil {
		return llm.ChatCompletionResponse{}, fmt.Errorf("chat with %s error: %w", req.Model, err)
	}
	return chatResp.ToChatCompletionResponse(), nil
}

func (p *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	reqBody := ChatRequest{}.FromChatCompletionRequest(req)
	chatResp, err := p.post(req.Model, reqBody, true)
	if err != nil {
		errChan <- fmt.Errorf("chat with %s error: %w", req.Model, err)
		return
	}
	dataChan <- chatResp.ToChatCompletionStreamResponse()
	errChan <- io.EOF
}

func (c *Client) post(model string, body ChatRequest, stream bool) (ChatResponse, error) {
	respBody := ChatResponse{}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return respBody, fmt.Errorf("marshal request body error: %w", err)
	}
	slog.Debug("request body", "body", string(reqBody))
	action := "generateContent"
	// if stream {
	// 	action = "streamGenerateContent"
	// }
	url := fmt.Sprintf("%s/v1beta/models/%s:%s?key=%s", defaultHost, model, action, c.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return respBody, fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.sess.Do(req)
	if err != nil {
		return respBody, fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var data any
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return respBody, fmt.Errorf("decode response error: %w", err)
		}
		return respBody, fmt.Errorf("response status code %d, error: %s", resp.StatusCode, data)
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return respBody, fmt.Errorf("decode response error: %w", err)
	}
	return respBody, nil
}
