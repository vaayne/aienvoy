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

	"github.com/Vaayne/aienvoy/pkg/llms/awsbedrock"
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/mitchellh/mapstructure"
)

type Client struct {
	session *http.Client
	config  llm.Config
}

func NewClient(cfg llm.Config) (*Client, error) {
	if cfg.LLMType != llm.LLMTypeAiGateway {
		return nil, fmt.Errorf("invalid config for ai gateway, llmtype: %s", cfg.LLMType)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client := &Client{
		session: http.DefaultClient,
		config:  cfg,
	}

	return client, nil
}

func (c *Client) ListModels() []string {
	return c.config.ListModels()
}

func (c *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	req.Stream = true
	config := c.config.AiGateway
	payload, err := buildRequestPayload(req, config)
	if err != nil {
		errChan <- fmt.Errorf("build request payload error: %w", err)
		return
	}

	url := config.GetChatURL(req.ModelId())
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
			switch config.Provider.Type {
			case llm.AiGatewayProviderAWSBedrock:
				var val awsbedrock.BedrockResponse
				if err := mapstructure.Decode(data, &val); err != nil {
					errChan <- fmt.Errorf("parse response error: %w", err)
					return
				}
				dataChan <- val.ToChatCompletionStreamResponse()
			case llm.AiGatewayProviderOpenAI, llm.AiGatewayProviderAzureOpenAI:
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

func buildRequestPayload(req llm.ChatCompletionRequest, config llm.AiGatewayConfig) ([]byte, error) {
	var payload []byte
	var err error

	switch config.Provider.Type {
	case llm.AiGatewayProviderAWSBedrock:
		bedrockRequest := &awsbedrock.BedrockRequest{}
		bedrockRequest.FromChatCompletionRequest(req)
		payload = bedrockRequest.Marshal()
	default:
		payload, _ = json.Marshal(req)
	}

	return payload, err
}

func setRequestHeaders(ctx context.Context, request *http.Request, config llm.AiGatewayConfig, stream bool, requestBody io.Reader) error {
	// set auth header
	for k, v := range config.GetAuthHeader() {
		request.Header.Set(k, v)
	}
	// set content type and accept
	request.Header.Set("Content-Type", "application/json")
	if stream {
		if config.Provider.Type == llm.AiGatewayProviderAWSBedrock {
			request.Header.Set("accept", "application/vnd.amazon.eventstream")
		} else {
			request.Header.Set("accept", "text/event-stream")
		}
	} else {
		request.Header.Set("accept", "application/json")
	}

	// set bedrock headers
	if config.Provider.Type == llm.AiGatewayProviderAWSBedrock {
		// Create a signer with the credentials
		signer := v4.NewSigner()
		h := sha256.New()
		_, _ = io.Copy(h, requestBody)
		payloadHash := hex.EncodeToString(h.Sum(nil))
		ab := config.Provider.AWSBedrock
		privider := aws.Credentials{AccessKeyID: ab.AccessKey, SecretAccessKey: ab.SecretKey, SessionToken: ""}
		if err := signer.SignHTTP(ctx, privider, request, payloadHash, "bedrock", ab.Region, time.Now()); err != nil {
			return fmt.Errorf("sign request error: %w", err)
		}
	}
	slog.DebugContext(ctx, "chat request headers", "headers", request.Header)
	return nil
}
