package together

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Vaayne/aienvoy/pkg/llms/llm"
	"github.com/Vaayne/aienvoy/pkg/llms/openai"
)

const baseUrl = "https://api.together.xyz"

var cacheModelsFile = filepath.Join(os.TempDir(), "together_models.json")

type Together *llm.LLM

func New(cfg llm.Config, dao llm.Dao) (Together, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return llm.New(dao, client), nil
}

type Client struct {
	*openai.Client
	config llm.Config
}

func NewClient(cfg llm.Config) (*Client, error) {
	if cfg.LLMType != llm.LLMTypeTogether {
		return nil, fmt.Errorf("invalid config, llmtype: %s", cfg.LLMType)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.BaseUrl == "" {
		cfg.BaseUrl = baseUrl
	}

	cli, err := openai.NewClient(cfg)

	return &Client{
		Client: cli,
		config: cfg,
	}, err
}

type Model struct {
	Id            string `json:"_id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	DisplayType   string `json:"display_type"`
	Description   string `json:"description"`
	ContextLength int    `json:"context_length"`
	Config        struct {
		ChatPrompt   string   `json:"chat_prompt"`
		PromptFormat string   `json:"prompt_format"`
		Stop         []string `json:"stop"`
	} `json:"config"`
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TogetherPythonOfficial/0.2.10")
}

func readModelsCache() []string {
	slog.Debug("read models cache", "file", cacheModelsFile)
	modelsFile, err := os.Open(cacheModelsFile)
	if err != nil {
		slog.Error("list models", "err", err)
		return []string{}
	}
	defer modelsFile.Close()
	var models []Model
	if err := json.NewDecoder(modelsFile).Decode(&models); err != nil {
		slog.Error("list models", "err", err)
		return []string{}
	}
	var modelNames []string
	for _, model := range models {
		if model.DisplayType == "chat" {
			modelNames = append(modelNames, model.Name)
		}
	}
	return modelNames
}

func saveModelsCache(models []any) {
	modelsFile, err := os.Create(cacheModelsFile)
	if err != nil {
		slog.Error("list models", "err", err)
		return
	}
	defer modelsFile.Close()
	if err := json.NewEncoder(modelsFile).Encode(models); err != nil {
		slog.Error("list models", "err", err)
		return
	}
}

func (c *Client) getModelsFromServer() []any {
	models := make([]any, 0)
	req, _ := http.NewRequest("GET", c.config.BaseUrl+"/models/info", nil)
	c.setHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("list models", "err", err)
		return models
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		slog.Error("list models", "err", err)
		return models
	}
	return models
}

func (c *Client) ListModels() []string {
	if _, err := os.Stat(cacheModelsFile); err != nil {
		models := c.getModelsFromServer()
		saveModelsCache(models)
	}
	return readModelsCache()
}
