package config

type Config struct {
	Service  ServiceConfig
	Admins   []Admin
	LLMs     []LLMConfig
	Axiom    Axiom
	Telegram struct {
		Token           string `yaml:"token"`
		ReadeaseChannel int64  `yaml:"readeaseChannel"`
	}
	ClaudeWeb struct {
		Token string `yaml:"token"`
	}
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
	ApiVersion  string
	Models      []string // valid model ids
}

type Axiom struct {
	Token   string
	Dataset string
}

type Admin struct {
	Email    string
	Password string
}
