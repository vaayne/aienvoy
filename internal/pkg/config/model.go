package config

type Config struct {
	Service ServiceConfig
	LLMs    []LLMConfig
	Axiom   Axiom
}

type ServiceConfig struct {
	Name     string
	Host     string
	Port     string
	URL      string
	LogLevel string
	Env      string
}

type LLMConfig struct {
	Type        string // openai or azureOpenAI
	ApiEndpoint string
	ApiKey      string
	Models      []string // valid model ids
}

type Axiom struct {
	Token   string
	Dataset string
}
