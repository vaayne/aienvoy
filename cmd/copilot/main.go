package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Vaayne/aienvoy/internal/pkg/parser"
	"github.com/Vaayne/aienvoy/pkg/llms/githubcopilot"
	"github.com/Vaayne/aienvoy/pkg/llms/llm"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appName               = "illm"
	defaultConfigFileName = "config.yaml"
	defaultModel          = "gpt-4"
)

var (
	llmClient *llm.LLM
	once      sync.Once
	llmConfig llm.Config
)

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: appName + " is a cli tool to run LLM",
	Long:  appName + ` is a tool to manage AI models in local or remote`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			slog.Error("prompt is required")
			os.Exit(1)
		}

		// init log
		if viper.GetBool("debug") {
			initLog(slog.LevelDebug)
		} else {
			initLog(slog.LevelInfo)
		}

		prompt := args[0]
		model := viper.GetString("model")
		if model == "" {
			model = defaultModel
		}
		if model == "" {
			slog.Error("model is required, you can set it in config file or use -m flag")
			os.Exit(1)
		}
		system := viper.GetString("system")
		files := viper.GetStringSlice("files")
		urls := viper.GetStringSlice("urls")
		texts := viper.GetStringSlice("texts")

		slog.Debug("start to run",
			"prompt", prompt, "model", model, "system", system,
			"files", files, "urls", urls, "texts", texts)
		ctx := context.Background()
		ctx, cancelFunc := context.WithTimeout(ctx, 300*time.Second)
		defer cancelFunc()
		chatStreaming(ctx, model, system, prompt, files, urls, texts)
	},
}

func setFlags() {
	bindFlag := func(flag string) {
		if err := viper.BindPFlag(flag, rootCmd.Flags().Lookup(flag)); err != nil {
			panic(err)
		}
	}

	rootCmd.Flags().BoolP("debug", "d", false, "log level")
	bindFlag("debug")

	rootCmd.Flags().StringSliceP("files", "f", []string{}, "more context from file, support multiple context files")
	bindFlag("files")
	rootCmd.Flags().StringSliceP("urls", "u", []string{}, "more context from url, support multiple context urls")
	bindFlag("urls")
	rootCmd.Flags().StringSliceP("texts", "t", []string{}, "more context from text, support multiple context texts")
	bindFlag("texts")
	rootCmd.Flags().StringP("system", "s", "", "system prompt")
	bindFlag("system")
	rootCmd.Flags().StringP("model", "m", "", "model")
	bindFlag("model")
	rootCmd.Flags().BoolP("help", "h", false, "help")
}

func init() {
	setFlags()
	initLLMConfig()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initLog(level slog.Level) {
	th := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	logger := slog.New(th)
	slog.SetDefault(logger)
}

type Config struct {
	DefaultModel string       `yaml:"default_model" mapstructure:"default_model"`
	LLMs         []llm.Config `yaml:"llms" mapstructure:"llms"`
}

func initLLMConfig() {
	// read file ~/.copilot-cli-access-token and set valeu to CopilotCLIAccessToken
	// Get the home directory
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("get home directory error", "err", err)
		os.Exit(1)
	}

	// Construct the file path
	filePath := filepath.Join(home, ".copilot-cli-access-token")

	f, err := os.Open(filePath)
	if err != nil {
		slog.Error("open file error", "err", err)
		os.Exit(1)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		slog.Error("read file error", "err", err)
		os.Exit(1)
	}
	token := strings.Trim(string(b), "\n")
	llmConfig = llm.Config{
		LLMType: llm.LLMTypeOpenAI,
		ApiKey:  token,
	}
}

func builsMessages(system, prompt string, files, urls, texts []string) ([]llm.ChatCompletionMessage, error) {
	messages := make([]llm.ChatCompletionMessage, 0)

	if prompt == "" {
		return nil, fmt.Errorf("prompt can't be empty")
	}

	if system != "" {
		messages = append(messages, llm.ChatCompletionMessage{
			Role:    llm.ChatMessageRoleSystem,
			Content: system,
		})
	}

	if len(files) > 0 {
		for _, f := range files {
			// the file may be a relative path, so we need to get the absolute path
			f, err := filepath.Abs(f)
			if err != nil {
				return nil, fmt.Errorf("get absolute path for file %s error: %w", f, err)
			}
			content, err := os.ReadFile(f)
			if err != nil {
				return nil, fmt.Errorf("read file %s error: %w", f, err)
			}
			messages = append(messages, llm.ChatCompletionMessage{
				Role:    llm.ChatMessageRoleUser,
				Content: fmt.Sprintf("\n\n-----\nHere is more context from file %s: \n\n%s\n-----\n\n", f, string(content)),
			})
		}
	}

	if len(urls) > 0 {
		p := parser.DefaultParser{}
		for _, u := range urls {
			content, err := p.Parse(u)
			if err != nil {
				return nil, fmt.Errorf("parse url %s error: %w", u, err)
			}
			messages = append(messages, llm.ChatCompletionMessage{
				Role:    llm.ChatMessageRoleUser,
				Content: fmt.Sprintf("\n\n-----\nHere is more context from url %s: \n\n%s\n-----\n\n", u, string(content.Content)),
			})
		}
	}

	if len(texts) > 0 {
		for _, t := range texts {
			messages = append(messages, llm.ChatCompletionMessage{
				Role:    llm.ChatMessageRoleUser,
				Content: fmt.Sprintf("\n\n-----\nHere is more context: \n\n%s\n-----\n\n", t),
			})
		}
	}

	messages = append(messages, llm.ChatCompletionMessage{
		Role:    llm.ChatMessageRoleUser,
		Content: prompt,
	})
	return messages, nil
}

func chatStreaming(ctx context.Context, model, system, prompt string, files, urls, texts []string) {

	once.Do(func() {
		var err error
		llmClient, err = githubcopilot.New(llmConfig, llm.DefaultDao)
		if err != nil {
			slog.Error("create llm service error", "err", err)
			os.Exit(1)
		}
	})

	messages, err := builsMessages(system, prompt, files, urls, texts)
	if err != nil {
		panic(err)
	}

	req := llm.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Stream:      true,
		Temperature: 0.9,
		MaxTokens:   4096,
		Stop:        []string{"</s>", "<|im_end|>"},
	}

	slog.Debug("start to create chat completion stream", "request", req)

	dataChan := make(chan llm.ChatCompletionStreamResponse)
	errChan := make(chan error)
	go llmClient.CreateChatCompletionStream(ctx, req, dataChan, errChan)
	for {
		select {
		case data := <-dataChan:
			if len(data.Choices) == 0 {
				continue
			}
			fmt.Print(data.Choices[0].Delta.Content)
		case err := <-errChan:
			if errors.Is(err, io.EOF) {
				fmt.Println()
				os.Exit(0)
			}
			slog.Error("error", "err", err)
			os.Exit(1)
		case <-ctx.Done():
			slog.Error("timeout")
			os.Exit(1)
		}
	}
}
