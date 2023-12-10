package aigateway

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/Vaayne/aienvoy/pkg/llm/claude"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/mitchellh/mapstructure"
)

func (c *Client) ListModels() []string {
	var models []string
	for _, config := range c.Mapping {
		models = append(models, config.Models...)
	}
	return models
}

func (c *Client) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	config, ok := c.Mapping[req.Model]
	if !ok {
		return llm.ChatCompletionResponse{}, fmt.Errorf("model id %s not found", req.Model)
	}

	payload, err := buildRequestPayload(req, config)
	if err != nil {
		return llm.ChatCompletionResponse{}, fmt.Errorf("build request payload error: %w", err)
	}

	url := config.GetBaseURL(req.Model)
	slog.DebugContext(ctx, "chat request", "url", url, "req", string(payload))
	requestBody := bytes.NewReader(payload)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBody)
	if err != nil {
		return llm.ChatCompletionResponse{}, fmt.Errorf("create request error: %w", err)
	}

	if err := setRequestHeaders(ctx, request, config, false, requestBody); err != nil {
		return llm.ChatCompletionResponse{}, fmt.Errorf("set request headers error: %w", err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return llm.ChatCompletionResponse{}, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	// check response status, if not 200, return error
	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return llm.ChatCompletionResponse{}, fmt.Errorf("decode response error: %w", err)
		}
		return llm.ChatCompletionResponse{}, fmt.Errorf("chat error, status: %s, body: %s, headers: %v", resp.Status, string(respBody), resp.Header)
	}
	return processResponse(ctx, resp.Body, config)
}

func (c *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	req.Stream = true
	config, ok := c.Mapping[req.Model]
	if !ok {
		errChan <- fmt.Errorf("model id %s not found", req.Model)
		return
	}
	payload, err := buildRequestPayload(req, config)
	if err != nil {
		errChan <- fmt.Errorf("build request payload error: %w", err)
		return
	}

	url := config.GetBaseURL(req.Model)
	slog.DebugContext(ctx, "chat request", "url", url, "req", string(payload))
	requestBody := bytes.NewReader(payload)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBody)
	if err != nil {
		errChan <- fmt.Errorf("create request error: %w", err)
		return
	}
	if err := setRequestHeaders(ctx, request, config, false, requestBody); err != nil {
		errChan <- fmt.Errorf("set request headers error: %w", err)
		return
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		errChan <- fmt.Errorf("do request error: %w", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			errChan <- fmt.Errorf("decode response error: %w", err)
			return
		}
		errChan <- fmt.Errorf("chat error, status: %s, body: %s, headers: %v", resp.Status, string(respBody), resp.Header)
		return
	}

	innerDataChan := make(chan any)
	innerErrChan := make(chan error)

	go llm.ParseSSE[any](resp.Body, innerDataChan, innerErrChan)
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-innerDataChan:
			switch config.Provider {
			case AWSBedrock:
				var val claude.BedrockResponse
				if err := mapstructure.Decode(data, &val); err != nil {
					errChan <- fmt.Errorf("parse response error: %w", err)
					return
				}
				dataChan <- val.ToChatCompletionStreamResponse()
			case OpenAI, AzureOpenAI:
				// convert any to ChatCompletionStreamResponse
				// the any may response as map[string]interface{}, so we have to convert it manually
				var val llm.ChatCompletionStreamResponse
				if err := mapstructure.Decode(data, &val); err != nil {
					errChan <- fmt.Errorf("parse response error: %w", err)
					return
				}
				dataChan <- val
			default:
				errChan <- fmt.Errorf("provider %s not supported", config.Provider)
			}
		case err := <-innerErrChan:
			errChan <- err
			return
		}
	}
}

func buildRequestPayload(req llm.ChatCompletionRequest, config Config) ([]byte, error) {
	var payload []byte
	var err error

	switch config.Provider {
	case AWSBedrock:
		bedrockRequest := &claude.BedrockRequest{}
		bedrockRequest.FromChatCompletionRequest(req)
		payload = bedrockRequest.Marshal()
	default:
		payload, _ = json.Marshal(req)
	}

	return payload, err
}

func setRequestHeaders(ctx context.Context, request *http.Request, config Config, stream bool, requestBody io.Reader) error {
	// set auth header
	for k, v := range config.GetAuthHeader() {
		request.Header.Set(k, v)
	}
	// set content type and accept
	request.Header.Set("Content-Type", "application/json")
	if stream {
		if config.Provider == AWSBedrock {
			request.Header.Set("accept", "application/vnd.amazon.eventstream")
		} else {
			request.Header.Set("accept", "text/event-stream")
		}
	} else {
		request.Header.Set("accept", "application/json")
	}

	// set bedrock headers
	if config.Provider == AWSBedrock {
		// Create a signer with the credentials
		signer := v4.NewSigner()
		h := sha256.New()
		_, _ = io.Copy(h, requestBody)
		payloadHash := hex.EncodeToString(h.Sum(nil))
		privider := aws.Credentials{AccessKeyID: config.AwsBedrockAccessKey, SecretAccessKey: config.AwsBedrockSecretKey, SessionToken: ""}
		if err := signer.SignHTTP(ctx, privider, request, payloadHash, "bedrock", config.AwsBedrockRegion, time.Now()); err != nil {
			return fmt.Errorf("sign request error: %w", err)
		}
	}
	return nil
}

func processResponse(ctx context.Context, body io.ReadCloser, config Config) (llm.ChatCompletionResponse, error) {
	var respBody llm.ChatCompletionResponse
	switch config.Provider {
	case AWSBedrock:
		var bedrockResp claude.BedrockResponse
		err := json.NewDecoder(body).Decode(&bedrockResp)
		if err != nil {
			return llm.ChatCompletionResponse{}, fmt.Errorf("decode response error: %w", err)
		}
		respBody = bedrockResp.ToChatCompletionResponse()
	default:
		err := json.NewDecoder(body).Decode(&respBody)
		if err != nil {
			return llm.ChatCompletionResponse{}, fmt.Errorf("decode response error: %w", err)
		}
	}
	slog.DebugContext(ctx, "chat response success", "resp", respBody)
	return respBody, nil
}
