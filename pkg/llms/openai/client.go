package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llms/llm"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	*openai.Client
	config llm.Config
}

var validLLMTypes = map[llm.LLMType]struct{}{
	llm.LLMTypeOpenAI:      {},
	llm.LLMTypeAzureOpenAI: {},
	llm.LLMTypeOpenRouter:  {},
	llm.LLMTypeTogether:    {},
	llm.LLMTypeAnyScale:    {},
}

func NewClient(cfg llm.Config) (*Client, error) {
	// make sure cfg.LLMType == llm.LLMTypeOpenAI
	// make sure cfg.ApiKey is not empty
	if _, ok := validLLMTypes[cfg.LLMType]; !ok {
		return nil, fmt.Errorf("invalid config for openai, llmtype: %s", cfg.LLMType)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var oaiConfig openai.ClientConfig
	if cfg.LLMType == llm.LLMTypeAzureOpenAI {
		baseUrl := fmt.Sprintf("https://%s.openai.azure.com", cfg.AzureOpenAI.ResourceName)
		oaiConfig = openai.DefaultAzureConfig(cfg.ApiKey, baseUrl)
		if cfg.AzureOpenAI.Version != "" {
			oaiConfig.APIVersion = cfg.AzureOpenAI.Version
		}
	} else {
		oaiConfig = openai.DefaultConfig(cfg.ApiKey)
		if cfg.BaseUrl != "" {
			oaiConfig.BaseURL = cfg.BaseUrl
		}
	}

	return &Client{
		Client: openai.NewClientWithConfig(oaiConfig),
		config: cfg,
	}, nil
}

func (s *Client) ListModels() []string {
	return s.config.ListModels()
}

func (s *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	openaiReq := toOpenAIChatCompletionRequest(req)
	slog.DebugContext(ctx, "chat start", "llm", openaiReq.Model, "is_stream", openaiReq.Stream)
	stream, err := s.Client.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		slog.InfoContext(ctx, "chat error", "llm", openaiReq.Model, "is_stream", openaiReq.Stream, "err", err)
		errChan <- err
		return
	}

	// log chat stream response
	sb := &strings.Builder{}
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.DebugContext(ctx, "chat success", "llm", openaiReq.Model, "is_stream", openaiReq.Stream)
				errChan <- err
				return
			}
			slog.InfoContext(ctx, "chat error", "llm", openaiReq.Model, "is_stream", openaiReq.Stream, "err", err)
			errChan <- err
			return
		}
		if len(resp.Choices) > 0 {
			sb.WriteString(resp.Choices[0].Delta.Content)
			dataChan <- toLLMChatCompletionStreamResponse(resp)
		}
	}
}
