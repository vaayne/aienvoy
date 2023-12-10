package together

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Vaayne/aienvoy/pkg/llm"
	llmconfig "github.com/Vaayne/aienvoy/pkg/llm/config"
)

const baseUrl = "https://api.together.xyz"

type Client struct {
	session *http.Client
	baseUrl string
	Apikey  string `json:"apikey"`
}

func NewClient(cfg llmconfig.Config) (*Client, error) {
	if cfg.LLMType != llmconfig.LLMTypeTogether {
		return nil, fmt.Errorf("invalid config for together, llmtype: %s", cfg.LLMType)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client := &Client{
		baseUrl: baseUrl,
		session: http.DefaultClient,
		Apikey:  cfg.ApiKey,
	}
	if cfg.BaseUrl != "" {
		client.baseUrl = cfg.BaseUrl
	}

	return client, nil
}

func (c *Client) WithSession(session *http.Client) *Client {
	c.session = session
	return c
}

func (c *Client) WithApikey(apikey string) *Client {
	c.Apikey = apikey
	return c
}

func (c *Client) WithBaseUrl(url string) *Client {
	c.baseUrl = url
	return c
}

type Model struct {
	Id            string `json:"_id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	DisplayType   string `json:"display_type"`
	Description   string `json:"description"`
	ContextLength int    `json:"context_length"`
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.Apikey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TogetherPythonOfficial/0.2.10")
}

func (c *Client) ListModels() []string {
	req, _ := http.NewRequest("GET", c.baseUrl+"/models/info", nil)
	c.setHeaders(req)
	resp, err := c.session.Do(req)
	if err != nil {
		slog.Error("list models", "err", err)
		return []string{}
	}
	defer resp.Body.Close()
	var models []Model
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		slog.Error("list models", "err", err)
		return []string{}
	}
	var modelNames []string
	for _, model := range models {
		modelNames = append(modelNames, model.Name)
	}
	return modelNames
}

func (c *Client) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	req.Stream = false
	togReq := &TogetherChatRequest{}
	togReq.FromChatCompletionRequest(req)
	reqBody, err := json.Marshal(togReq)
	if err != nil {
		return llm.ChatCompletionResponse{}, fmt.Errorf("create chat completion marshal request error: %w", err)
	}

	httpReq, _ := http.NewRequest("POST", c.baseUrl+"/v1/completions", bytes.NewBuffer(reqBody))
	c.setHeaders(httpReq)
	resp, err := c.session.Do(httpReq)
	if err != nil {
		return llm.ChatCompletionResponse{}, fmt.Errorf("create chat completion error: %w", err)
	}
	defer resp.Body.Close()
	var togResp TogetherChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&togResp); err != nil {
		return llm.ChatCompletionResponse{}, fmt.Errorf("create chat completion decode response error: %w", err)
	}

	return togResp.ToChatCompletionResponse(), nil
}

func (c *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	req.Stream = true
	togReq := &TogetherChatRequest{}
	togReq.FromChatCompletionRequest(req)
	reqBody, err := json.Marshal(togReq)
	if err != nil {
		errChan <- fmt.Errorf("create chat completion stream marshal request error: %w", err)
		return
	}

	httpReq, _ := http.NewRequest("POST", c.baseUrl+"/v1/completions", bytes.NewBuffer(reqBody))
	c.setHeaders(httpReq)
	resp, err := c.session.Do(httpReq)
	if err != nil {
		errChan <- fmt.Errorf("create chat completion stream error: %w", err)
		return
	}
	defer resp.Body.Close()

	innerDataChan := make(chan TogetherChatResponse)
	defer close(innerDataChan)
	innerErrChan := make(chan error)
	defer close(innerErrChan)

	go llm.ParseSSE(resp.Body, innerDataChan, innerErrChan)
	for {
		select {
		case data := <-innerDataChan:
			dataChan <- data.ToChatCompletionStreamResponse()
		case err := <-innerErrChan:
			errChan <- err
			return
		}
	}
}
