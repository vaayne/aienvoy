package config

import "fmt"

type LLMType string

func (t LLMType) String() string {
	return string(t)
}

const (
	LLMTypeOpenAI      LLMType = "openai"
	LLMTypeAiGateway   LLMType = "aigateway"
	LLMTypeAzureOpenAI LLMType = "azure-openai"
	LLMTypeAWSBedrock  LLMType = "aws-bedrock"
	LLMTypeTogether    LLMType = "together"
	LLMTypeReplicate   LLMType = "replicate"
	LLMTypeClaudeWeb   LLMType = "claude-web"
	LLMTypeGoogleBard  LLMType = "google-bard"
)

type AiGatewayProviderType string

func (p AiGatewayProviderType) String() string {
	return string(p)
}

// create constants for the different AI providers
// AiGatewayProviderWorkersAI, OpenAI, HuggingFace, Replicate, AzureOpenAI, AWSBedrock
const (
	AiGatewayProviderWorkersAI   AiGatewayProviderType = "workers-ai"
	AiGatewayProviderOpenAI      AiGatewayProviderType = "openai"
	AiGatewayProviderHuggingFace AiGatewayProviderType = "huggingface"
	AiGatewayProviderReplicate   AiGatewayProviderType = "replicate"
	AiGatewayProviderAzureOpenAI AiGatewayProviderType = "azure-openai"
	AiGatewayProviderAWSBedrock  AiGatewayProviderType = "aws-bedrock"
)

const AIGatewayHost = "https://gateway.ai.cloudflare.com/v1"

type Config struct {
	// LLMType is the type of LLM to use
	LLMType LLMType `json:"type" yaml:"type" mapstructure:"type"`
	// Models is a list of valid model ids for this config
	Models []string `json:"models" yaml:"models" mapstructure:"models"`

	// ApiKey is the API key for the provider, works for OpenAI, HuggingFace, Replicate and Together
	ApiKey string `json:"api_key" yaml:"api_key" mapstructure:"api_key"`
	// BaseUrl is the base url for the provider, works for OpenAI, HuggingFace, Replicate and Together
	BaseUrl string `json:"base_url" yaml:"base_url" mapstructure:"base_url"`

	// AzureOpenAI is the config for Azure OpenAI
	AzureOpenAI AzureOpenAIConfig `json:"azure_openai" yaml:"azure_openai" mapstructure:"azure_openai"`
	// AWSBedrock is the config for AWS Bedrock
	AWSBedrock AWSBedrockConfig `json:"aws_bedrock" yaml:"aws_bedrock" mapstructure:"aws_bedrock"`
	// AiGateway is the config for Cloudflare AI Gateway
	AiGateway AiGatewayConfig `json:"aigateway" yaml:"aigateway" mapstructure:"aigateway"`
}

func (c Config) Validate() error {
	if c.LLMType == "" {
		return fmt.Errorf("llm.type is required")
	}

	switch c.LLMType {
	case LLMTypeOpenAI, LLMTypeClaudeWeb, LLMTypeGoogleBard, LLMTypeTogether, LLMTypeReplicate:
		if c.ApiKey == "" {
			return fmt.Errorf("api_key is required")
		}
	case LLMTypeAzureOpenAI:
		return c.AzureOpenAI.validate()
	case LLMTypeAWSBedrock:
		return c.AWSBedrock.validate()
	case LLMTypeAiGateway:
		return c.AiGateway.validate()
	}
	switch c.LLMType {
	case LLMTypeAzureOpenAI:
		if err := c.AzureOpenAI.validate(); err != nil {
			return err
		}
	case LLMTypeAWSBedrock:
		if err := c.AWSBedrock.validate(); err != nil {
			return err
		}
	case LLMTypeAiGateway:
		if err := c.AiGateway.validate(); err != nil {
			return err
		}
	}
	return nil
}

type AzureOpenAIConfig struct {
	ApiKey       string            `json:"api_key" mapstructure:"api_key" yaml:"api_key"`
	ResourceName string            `json:"resource_name" mapstructure:"resource_name"`
	ModelMapping map[string]string `json:"model_mapping" mapstructure:"model_mapping"`
	Version      string            `json:"version" mapstructure:"version"`
}

func (c *AzureOpenAIConfig) validate() error {
	if c.ApiKey == "" {
		return fmt.Errorf("azure_openai.api_key is required")
	}
	if c.ResourceName == "" {
		return fmt.Errorf("azure_openai.resource_name is required")
	}
	return nil
}

type AWSBedrockConfig struct {
	// AccessKey is the access key for AWS Bedrock
	AccessKey string `json:"access_key" mapstructure:"access_key" yaml:"access_key"`
	// SecretKey is the secret key for AWS Bedrock
	SecretKey string `json:"secret_key" mapstructure:"secret_key"`
	// egion is the region for AWS Bedrock
	Region string `json:"region" mapstructure:"region"`
}

func (c *AWSBedrockConfig) validate() error {
	if c.AccessKey == "" {
		return fmt.Errorf("aws_bedrock.access_key is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("aws_bedrock.secret_key is required")
	}
	if c.Region == "" {
		return fmt.Errorf("aws_bedrock.region is required")
	}
	return nil
}

type AiGatewayProvider struct {
	Type        AiGatewayProviderType `json:"type" mapstructure:"type" yaml:"type"`
	ApiKey      string                `json:"api_key" mapstructure:"api_key" yaml:"api_key"`
	AzureOpenAI AzureOpenAIConfig     `json:"azure_openai" mapstructure:"azure_openai" yaml:"azure_openai"`
	AWSBedrock  AWSBedrockConfig      `json:"aws_bedrock" mapstructure:"aws_bedrock" yaml:"aws_bedrock"`
}

type AiGatewayConfig struct {
	// AccountId is the account tag for the AI Gateway
	AccountId string `json:"account_id" mapstructure:"account_id" yaml:"account_id"`
	// Name is the name of the gateway
	Name string `json:"name" mapstructure:"name" yaml:"name"`
	// Provider is the provider type of AI Gateway
	Provider AiGatewayProvider `json:"provider" mapstructure:"provider" yaml:"provider"`
}

func (c *AiGatewayConfig) validate() error {
	if c.AccountId == "" {
		return fmt.Errorf("aigateway.account_id is required")
	}
	if c.Name == "" {
		return fmt.Errorf("aigateway.name is required")
	}
	if c.Provider.Type == "" {
		return fmt.Errorf("aigateway.provider.type is required")
	}
	switch c.Provider.Type {
	case AiGatewayProviderOpenAI, AiGatewayProviderHuggingFace, AiGatewayProviderReplicate:
		if c.Provider.ApiKey == "" {
			return fmt.Errorf("aigateway.provider.api_key is required")
		}
	case AiGatewayProviderAzureOpenAI:
		return c.Provider.AzureOpenAI.validate()
	case AiGatewayProviderAWSBedrock:
		return c.Provider.AWSBedrock.validate()
	}
	return nil
}

func (c *AiGatewayConfig) GetChatURL(model string) string {
	baseUrl := fmt.Sprintf("%s/%s/%s/%s", AIGatewayHost, c.AccountId, c.Name, c.Provider.Type)
	switch c.Provider.Type {
	case AiGatewayProviderOpenAI, AiGatewayProviderHuggingFace, AiGatewayProviderReplicate:
		return fmt.Sprintf("%s/chat/completions", baseUrl)
	case AiGatewayProviderWorkersAI:
		return fmt.Sprintf("%s/%s", baseUrl, model)
	case AiGatewayProviderAzureOpenAI:
		az := c.Provider.AzureOpenAI
		return fmt.Sprintf("%s/%s/%s/chat/completions?api-version=%s", baseUrl, az.ResourceName, az.ModelMapping[model], az.Version)
	case AiGatewayProviderAWSBedrock:
		ab := c.Provider.AWSBedrock
		return fmt.Sprintf("%s/bedrock-runtime/%s/model/%s/invoke", baseUrl, ab.Region, model)
	}
	return ""
}

func (c AiGatewayConfig) GetAuthHeader() map[string]string {
	switch c.Provider.Type {
	case AiGatewayProviderOpenAI, AiGatewayProviderHuggingFace, AiGatewayProviderWorkersAI:
		return map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", c.Provider.ApiKey),
		}
	case AiGatewayProviderReplicate:
		return map[string]string{
			"Authorization": fmt.Sprintf("Token %s", c.Provider.ApiKey),
		}
	case AiGatewayProviderAzureOpenAI:
		return map[string]string{
			"api-key": c.Provider.ApiKey,
		}
	case AiGatewayProviderAWSBedrock:
		ab := c.Provider.AWSBedrock
		return map[string]string{
			"Authorization": fmt.Sprintf("Basic %s %s", ab.AccessKey, ab.SecretKey),
		}
	}
	return nil
}
