package aigateway

import (
	"fmt"
)

type Provider string

func (p Provider) String() string {
	return string(p)
}

// create constants for the different AI providers
// WorkersAI, OpenAI, HuggingFace, Replicate, AzureOpenAI, AWSBedrock
const (
	WorkersAI   Provider = "workers-ai"
	OpenAI      Provider = "openai"
	HuggingFace Provider = "huggingface"
	Replicate   Provider = "replicate"
	AzureOpenAI Provider = "azure-openai"
	AWSBedrock  Provider = "aws-bedrock"
)

const AIGatewayHost = "https://gateway.ai.cloudflare.com/v1"

var defaultOpenAIModels = []string{
	"gpt-3.5-turbo", "gpt-3.5-turbo-1106", "gpt-3.5-turbo-16k",
	"gpt-4-1106-preview", "gpt-4-vision-preview", "gpt-4", "gpt-4-32k", "gpt-4-0613", "gpt-4-32k-0613",
}

var defaultWorkersAIModels = []string{
	"@cf/meta/llama-2-7b-chat-fp16", "@cf/meta/llama-2-7b-chat-int8",
	"@cf/mistral/mistral-7b-instruct-v0.1", "@hf/thebloke/codellama-7b-instruct-awq",
}

var defaultAwsBedrockModels = []string{"anthropic.claude-v1", "anthropic.claude-v2", "anthropic.claude-v2:1", "anthropic.claude-instant-v1"}

type Config struct {
	// AccountTag is the account tag for the AI Gateway
	AccountTag string `json:"account_tag"`
	// GatewayName is the name of the gateway
	GatewayName string `json:"gateway_name"`
	// Provider is the provider type of AI Gateway
	Provider Provider `json:"provider"`
	// ApiKey is the API key for the provider
	ApiKey string `json:"api_key"`
	// Models is a list of valid model ids for this config
	Models []string `json:"models"`
	// ResourceName is the name of the resource group, without the .openai.azure.com
	// e.g. "openai-rg"
	AzureOpenAIResourceName string `json:"azure_openai_resource_name"`
	// ModelMapping is a map of model id to deployed model name
	// e.g. {"gpt-3.5-trurbo": "gpt-35", "gpt-4-turbo": "gpt-4"}
	AzureModelMapping map[string]string
	// AzureVersion is the version of the Azure API to use
	AzureVersion string `json:"azure_version"`
	// AWSBedrockAccessKey is the access key for AWS Bedrock
	AwsBedrockAccessKey string `json:"aws_bedrock_access_key"`
	// AWSBedrockSecretKey is the secret key for AWS Bedrock
	AwsBedrockSecretKey string `json:"aws_bedrock_secret_key"`
	// AWSBedrockRegion is the region for AWS Bedrock
	AwsBedrockRegion string `json:"aws_bedrock_region"`
}

func (c *Config) GetBaseURL(model string) string {
	baseUrl := fmt.Sprintf("%s/%s/%s/%s", AIGatewayHost, c.AccountTag, c.GatewayName, c.Provider)
	switch c.Provider {
	case OpenAI, HuggingFace, Replicate:
		return fmt.Sprintf("%s/chat/completions", baseUrl)
	case WorkersAI:
		return fmt.Sprintf("%s/%s", baseUrl, model)
	case AzureOpenAI:
		return fmt.Sprintf("%s/%s/%s/chat/completions?api-version=%s", baseUrl, c.AzureOpenAIResourceName, c.AzureModelMapping[model], c.AzureVersion)
	case AWSBedrock:
		return fmt.Sprintf("%s/bedrock-runtime/%s/model/%s/invoke", baseUrl, c.AwsBedrockRegion, model)
	}
	return ""
}

func (c Config) GetAuthHeader() map[string]string {
	switch c.Provider {
	case OpenAI, HuggingFace, WorkersAI:
		return map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", c.ApiKey),
		}
	case Replicate:
		return map[string]string{
			"Authorization": fmt.Sprintf("Token %s", c.ApiKey),
		}
	case AzureOpenAI:
		return map[string]string{
			"api-key": c.ApiKey,
		}
	case AWSBedrock:
		return map[string]string{
			"Authorization": fmt.Sprintf("Basic %s %s", c.AwsBedrockAccessKey, c.AwsBedrockSecretKey),
		}
	}
	return nil
}

type Client struct {
	Mapping map[string]Config `json:"mapping"`
}

func New(configs ...Config) (*Client, error) {
	client := &Client{
		Mapping: make(map[string]Config),
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("no configs provided")
	}
	for _, config := range configs {
		if config.Models == nil || len(config.Models) == 0 {
			switch config.Provider {
			case OpenAI:
				config.Models = defaultOpenAIModels
			case WorkersAI:
				config.Models = defaultWorkersAIModels
			case AWSBedrock:
				config.Models = defaultAwsBedrockModels
			case AzureOpenAI:
				if config.AzureModelMapping == nil || len(config.AzureModelMapping) == 0 {
					return nil, fmt.Errorf("azure model mapping must be provided")
				}
				config.Models = make([]string, 0, len(config.AzureModelMapping))
				for model := range config.AzureModelMapping {
					config.Models = append(config.Models, model)
				}
			default:
				if config.Models == nil || len(config.Models) == 0 {
					return nil, fmt.Errorf("must provide valid models for provider %s", config.Provider)
				}
			}
		}

		for _, model := range config.Models {
			client.Mapping[model] = config
		}
	}
	return client, nil
}
