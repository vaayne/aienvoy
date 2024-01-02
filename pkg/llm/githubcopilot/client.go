package githubcopilot

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Vaayne/aienvoy/pkg/cache"
	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/Vaayne/aienvoy/pkg/session"
)

const (
	copilotTokenURL = "https://api.github.com/copilot_internal/v2/token"
	defaultChatURL  = "https://api.githubcopilot.com/chat/completions"
)

type Client struct {
	session *session.Session
	apiKey  string
	chatUrl string
}

func NewClient(apiKey string) *Client {
	return &Client{
		session: session.New(),
		apiKey:  apiKey,
		chatUrl: defaultChatURL,
	}
}

func (c *Client) ListModels() []string {
	return []string{"gpt-3.5-turbo", "gpt-4"}
}

func (c *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	cReq := &Request{}
	cReq.FromChatCompletionRequest(req)
	body, _ := json.Marshal(cReq)

	hReq, err := http.NewRequest(http.MethodPost, c.chatUrl, bytes.NewReader(body))
	if err != nil {
		errChan <- fmt.Errorf("create chat completion stream error: %w", err)
		return
	}

	copilotToken, err := c.getCopilotToken(c.apiKey)
	if err != nil {
		errChan <- err
		return
	}
	hReq.Header = c.buildHeaders(copilotToken)
	resp, err := c.session.Do(hReq)
	if err != nil {
		errChan <- fmt.Errorf("create chat completion stream error: %w", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("copilot response error: %s", resp.Status)
		return
	}

	llm.ParseSSE(resp.Body, dataChan, errChan)
}

func (c *Client) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	dataChan := make(chan llm.ChatCompletionStreamResponse)
	defer close(dataChan)
	errChan := make(chan error)
	defer close(errChan)

	go c.CreateChatCompletionStream(ctx, req, dataChan, errChan)
	sb := strings.Builder{}
	resp := llm.ChatCompletionResponse{}

	for {
		select {
		case data := <-dataChan:
			if len(data.Choices) == 0 {
				continue
			}
			resp = data.ToChatCompletionResponse()
		case err := <-errChan:
			if errors.Is(err, io.EOF) {
				resp.Choices[0].Message.Content = sb.String()
				return resp, nil
			}
			slog.Error("\nerr", "err", err)
			return llm.ChatCompletionResponse{}, err
		}
	}
}

// getCopilotToken retrieves a token for GitHub Copilot.
func (c *Client) getCopilotToken(githubToken string) (string, error) {
	// Try to get the token from the cache.
	token, ok := cache.DefaultClient.Get(githubToken)
	if ok {
		// If the token is in the cache, return it.
		return token.(string), nil
	}

	// Create a new HTTP request to get the Copilot token.
	req, _ := http.NewRequest("GET", copilotTokenURL, nil)
	// Set the Authorization header with the GitHub token.
	req.Header.Set("Authorization", "token "+githubToken)
	// Send the request.
	response, err := c.session.Do(req)
	if err != nil || response.StatusCode != 200 {
		// If there's an error or the status code is not 200, return an error.
		return "", fmt.Errorf("get copilot token error: %w", err)
	}
	// Ensure the response body is closed after the function returns.
	defer response.Body.Close()
	// Read the response body.
	body, _ := io.ReadAll(response.Body)
	// Define a struct to hold the Copilot token and its expiration time.
	var copilotToken struct {
		Token     string `json:"token"`
		ExpiresAt int64  `json:"expires_at"`
	}

	// Unmarshal the JSON response into the copilotToken struct.
	if err = json.Unmarshal(body, &copilotToken); err != nil {
		// If there's an error unmarshalling, return an error.
		return "", fmt.Errorf("get copilot token error: %w", err)
	}
	// Convert the expiration time from Unix timestamp to time.Time, subtracting 500 seconds as a buffer.
	expirationTime := time.Unix(copilotToken.ExpiresAt-500, 0)
	// Set the token in the cache, with the expiration time.
	cache.DefaultClient.Set(githubToken, copilotToken.Token, time.Until(expirationTime))
	// Return the Copilot token.
	return copilotToken.Token, nil
}

func (c *Client) buildHeaders(copilotToken string) http.Header {
	headers := http.Header{}

	generateHexStr := func(length int) string {
		bytes := make([]byte, length/2)
		if _, err := rand.Read(bytes); err != nil {
			panic(err)
		}
		return hex.EncodeToString(bytes)
	}

	headers.Set("Authorization", "Bearer "+copilotToken)
	headers.Set("X-Request-Id", generateHexStr(8)+"-"+generateHexStr(4)+"-"+generateHexStr(4)+"-"+generateHexStr(4)+"-"+generateHexStr(12))
	headers.Set("Vscode-Sessionid", generateHexStr(8)+"-"+generateHexStr(4)+"-"+generateHexStr(4)+"-"+generateHexStr(4)+"-"+generateHexStr(25))
	headers.Set("Vscode-Machineid", generateHexStr(64))
	headers.Set("Editor-Version", "vscode/1.83.1")
	headers.Set("Editor-Plugin-Version", "copilot-chat/0.8.0")
	headers.Set("Openai-Organization", "github-copilot")
	headers.Set("Openai-Intent", "conversation-panel")
	headers.Set("Content-Type", "text/event-stream; charset=utf-8")
	headers.Set("User-Agent", "GitHubCopilotChat/0.8.0")
	headers.Set("Accept", "*/*")
	headers.Set("Accept-Encoding", "gzip,deflate,br")
	headers.Set("Connection", "close")
	return headers
}
