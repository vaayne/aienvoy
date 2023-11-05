package llmservice

import (
	"context"
	"log/slog"
	"time"

	"github.com/Vaayne/aienvoy/internal/pkg/cache"
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/cookiecloud"
	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/Vaayne/aienvoy/pkg/llm/bard"
	"github.com/Vaayne/aienvoy/pkg/llm/claude"
	"github.com/Vaayne/aienvoy/pkg/llm/claudeweb"
	"github.com/Vaayne/aienvoy/pkg/llm/openai"
	"github.com/Vaayne/aienvoy/pkg/llm/phind"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	goopenai "github.com/sashabaranov/go-openai"
)

type Service interface {
	ListModels() []string
	CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error)
	CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, respChan chan llm.ChatCompletionStreamResponse, errChan chan error)

	CreateConversation(ctx context.Context, name string) (llm.Conversation, error)
	ListConversations(ctx context.Context) ([]llm.Conversation, error)
	GetConversation(ctx context.Context, id string) (llm.Conversation, error)
	DeleteConversation(ctx context.Context, id string) error

	CreateMessageStream(ctx context.Context, conversationId string, req llm.ChatCompletionRequest, respChan chan llm.ChatCompletionStreamResponse, errChan chan error)
	CreateMessage(ctx context.Context, conversationId string, req llm.ChatCompletionRequest) (llm.Message, error)
	ListMessages(ctx context.Context, conversationId string) ([]llm.Message, error)
	GetMessage(ctx context.Context, id string) (llm.Message, error)
	DeleteMessage(ctx context.Context, id string) error
}

const (
	modelClientCacheKeyPrefix = "llm:"
	llmTypeAzureOpenAI        = "azure-openai"
	llmTypeOpenAI             = "openai"
)

func New(model string, dao llm.Dao) Service {
	cachekey := modelClientCacheKeyPrefix + model

	client, ok := cache.DefaultClient.Get(cachekey)
	if ok {
		return client.(Service)
	}
	var cli Service
	switch model {
	case phind.ModelPhindV1:
		cli = newPhind(dao)
	case openai.ModelGPT3Dot5Turbo:
		cli = newOpenai(dao)
	case claudeweb.ModelClaudeWeb:
		cli = newClaudeWeb(dao)
	case claude.ModelClaudeInstantV1Dot2:
		cli = newClaude(dao)
	case bard.ModelBard:
		cli = newBard(dao)
	default:
		slog.Error("unsupported model: " + model)
	}

	cache.DefaultClient.Set(cachekey, cli, 5*time.Minute)
	return cli
}

func newPhind(dao llm.Dao) *phind.Phind {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	cookies, err := cc.GetHttpCookies("www.phind.com")
	if err != nil {
		slog.Error("get cookies error", "err", err, "domain", "www.phind.com")
		return nil
	}

	cookies1, err := cc.GetHttpCookies(".phind.com")
	if err != nil {
		slog.Error("get cookies error", "err", err, "domain", ".phind.com")
		return nil
	}
	cookies = append(cookies, cookies1...)
	return phind.New(cookies, dao)
}

func newOpenai(dao llm.Dao) *openai.OpenAI {
	var clientCfg goopenai.ClientConfig
	cfg := config.GetConfig().LLMs[0]
	if cfg.Type == llmTypeAzureOpenAI {
		clientCfg = goopenai.DefaultAzureConfig(cfg.ApiKey, cfg.ApiEndpoint)
		if cfg.ApiVersion != "" {
			clientCfg.APIVersion = cfg.ApiVersion
		}
	} else if cfg.Type == llmTypeOpenAI {
		clientCfg = goopenai.DefaultConfig(cfg.ApiKey)
		if cfg.ApiEndpoint != "" {
			clientCfg.BaseURL = cfg.ApiEndpoint
		}
	} else {
		slog.Error("unknown LLM type", "type", cfg.Type)
		return nil
	}

	return openai.New(clientCfg, dao)
}

func newClaudeWeb(dao llm.Dao) *claudeweb.ClaudeWeb {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	sessionKey, err := cc.GetCookie("claude.ai", "sessionKey")
	if err != nil {
		slog.Error("get cookie error", "err", err)
		return nil
	}

	return claudeweb.New(sessionKey.Value, dao)
}

func newClaude(dao llm.Dao) *claude.Claude {
	getAWSConfig := func() aws.Config {
		cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(config.GetConfig().AWS.Region),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				config.GetConfig().AWS.AccessKeyId,
				config.GetConfig().AWS.SecretAccessKey,
				"",
			)))
		if err != nil {
			slog.Error("get aws config error", "err", err)
			return aws.Config{}
		}
		return cfg
	}
	return claude.New(getAWSConfig(), dao)
}

func newBard(dao llm.Dao) *bard.Bard {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	getCookie := func(key string) string {
		val, err := cc.GetCookie(".google.com", key)
		if err != nil {
			slog.Error("get cookie error", "err", err)
			return ""
		}
		return val.Value
	}

	cli, err := bard.New(getCookie("__Secure-1PSID"), dao, bard.WithCookies(map[string]string{
		"__Secure-1PSID":   getCookie("__Secure-1PSID"),
		"__Secure-1PSIDCC": getCookie("__Secure-1PSIDCC"),
		"__Secure-1PSIDTS": getCookie("__Secure-1PSIDTS"),
	}), bard.WithTimeout(120*time.Second))
	if err != nil {
		slog.Error("init bard client error", "err", err)
		return nil
	}
	return cli
}
