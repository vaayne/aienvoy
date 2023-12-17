package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llm"
	llmconfig "github.com/Vaayne/aienvoy/pkg/llm/config"
	"github.com/sashabaranov/go-openai"
)

type Client struct {
	*openai.Client
	Models []string `json:"models" yaml:"models"`
	config llmconfig.Config
}

func NewClient(cfg llmconfig.Config) (*Client, error) {
	// make sure cfg.LLMType == llmconfig.LLMTypeOpenAI
	// make sure cfg.ApiKey is not empty
	if cfg.LLMType != llmconfig.LLMTypeOpenAI && cfg.LLMType != llmconfig.LLMTypeAzureOpenAI {
		return nil, fmt.Errorf("invalid config for openai, llmtype: %s", cfg.LLMType)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var oaiConfig openai.ClientConfig
	if cfg.LLMType == llmconfig.LLMTypeOpenAI {
		oaiConfig = openai.DefaultConfig(cfg.ApiKey)
		if cfg.BaseUrl != "" {
			oaiConfig.BaseURL = cfg.BaseUrl
		}
	} else if cfg.LLMType == llmconfig.LLMTypeAzureOpenAI {
		baseUrl := fmt.Sprintf("https://%s.openai.azure.com", cfg.AzureOpenAI.ResourceName)
		oaiConfig = openai.DefaultAzureConfig(cfg.ApiKey, baseUrl)
		if cfg.AzureOpenAI.Version != "" {
			oaiConfig.APIVersion = cfg.AzureOpenAI.Version
		}
	}

	return &Client{
		Client: openai.NewClientWithConfig(oaiConfig),
		config: cfg,
		Models: cfg.Models,
	}, nil
}

func (s *Client) ListModels() []string {
	if len(s.Models) == 0 {
		if s.config.LLMType == llmconfig.LLMTypeAzureOpenAI {
			s.Models = llmconfig.DefaultOpenAIChatModels
		} else {
			s.Models = s.config.AzureOpenAI.ListModels()
		}
	}
	return s.Models
}

func (s *Client) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	openaiReq := toOpenAIChatCompletionRequest(req)
	slog.DebugContext(ctx, "chat start", "llm", openaiReq.Model, "is_stream", openaiReq.Stream)
	resp, err := s.Client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		slog.ErrorContext(ctx, "chat with OpenAI error", "err", err)
		return llm.ChatCompletionResponse{}, err
	}
	slog.DebugContext(ctx, "chat success", "llm", openaiReq.Model, "is_stream", openaiReq.Stream)

	return toLLMChatCompletionResponse(resp), nil
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
