package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Vaayne/aienvoy/internal/core/llm"
	"github.com/Vaayne/aienvoy/internal/core/llm/llmclaude"

	"github.com/sashabaranov/go-openai"

	"github.com/Vaayne/aienvoy/internal/core/llm/llmopenai"

	"github.com/labstack/echo/v5"
)

type LLMHandler struct {
	openai *llmopenai.OpenAI
	claude *llmclaude.Claude
}

func NewLLMHandler() *LLMHandler {
	return &LLMHandler{
		openai: llmopenai.New(),
		claude: llmclaude.New(),
	}
}

func (l *LLMHandler) getService(model string) llm.Service {
	switch model {
	// openai base model
	case openai.GPT3Dot5Turbo, openai.GPT3Dot5Turbo16K, openai.GPT4, openai.GPT432K, openai.GPT3Dot5TurboInstruct:
		return l.openai
	// openai time limited model
	case openai.GPT3Dot5Turbo0301, openai.GPT3Dot5Turbo0613, openai.GPT3Dot5Turbo16K0613, openai.GPT40314, openai.GPT40613, openai.GPT432K0314, openai.GPT432K0613:
		return l.openai
	// claude models
	case llmclaude.ModelClaudeV2, llmclaude.ModelClaudeV1Dot3, llmclaude.ModelClaudeInstantV1Dot2:
		return l.claude
	default:
		slog.Error("unknown model", "model", model)
		return nil
	}
}

func (l *LLMHandler) GetModels(c echo.Context) error {
	resp, err := l.openai.GetModels(c.Request().Context())
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "get models error", "err", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (l *LLMHandler) CreateChatCompletion(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(openai.ChatCompletionRequest)
	err := c.Bind(req)
	if err != nil {
		slog.ErrorContext(ctx, "bind chat request body error", "err", err.Error())
		return c.String(http.StatusBadRequest, "bad request")
	}

	svc := l.getService(req.Model)
	if svc == nil {
		return c.String(http.StatusBadRequest, "unknown model")
	}

	if req.Stream {
		return l.chatStream(c, svc, req)
	}

	resp, err := svc.CreateChatCompletion(ctx, req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (l *LLMHandler) CreateCompletion(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(openai.CompletionRequest)
	err := c.Bind(req)
	if err != nil {
		slog.ErrorContext(ctx, "bind chat request body error", "err", err.Error())
		return c.String(http.StatusBadRequest, "bad request")
	}

	svc := l.getService(req.Model)
	if svc == nil {
		return c.String(http.StatusBadRequest, "unknown model")
	}

	if req.Stream {
		return l.completionStream(c, svc, req)
	}

	resp, err := svc.CreateCompletion(ctx, req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (l *LLMHandler) CreateEmbeddings(c echo.Context) error {
	var req *openai.EmbeddingRequest
	err := c.Bind(req)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "bind embedding request body error", "err", err.Error())
		return c.String(http.StatusBadRequest, "bad request")
	}

	resp, err := l.openai.CreateEmbeddings(c.Request().Context(), *req)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "create embedding error", "err", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (l *LLMHandler) chatStream(c echo.Context, svc llm.Service, req *openai.ChatCompletionRequest) error {
	dataChan := make(chan openai.ChatCompletionStreamResponse)
	defer close(dataChan)
	errChan := make(chan error)
	defer close(errChan)

	go svc.CreateChatCompletionStream(c.Request().Context(), req, dataChan, errChan)

	// sse stream response
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	c.Response().WriteHeader(http.StatusOK)

	for {
		select {
		case data := <-dataChan:
			msg, err := json.Marshal(data)
			if err != nil {
				slog.ErrorContext(c.Request().Context(), "chat stream marshal response error", "err", err.Error())
				return c.String(http.StatusInternalServerError, err.Error())
			}
			_, err = c.Response().Write([]byte(fmt.Sprintf("data: %s\n\n", msg)))
			if err != nil {
				slog.ErrorContext(c.Request().Context(), "write chat stream response error", "err", err.Error())
				return c.String(http.StatusInternalServerError, err.Error())
			}
			c.Response().Flush()
		case err := <-errChan:
			if errors.Is(err, io.EOF) {
				return c.String(http.StatusOK, "data: [DONE]\n\n")
			}
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
}

func (l *LLMHandler) completionStream(c echo.Context, svc llm.Service, req *openai.CompletionRequest) error {
	dataChan := make(chan openai.CompletionResponse)
	defer close(dataChan)
	errChan := make(chan error)
	defer close(errChan)

	go svc.CreateCompletionStream(c.Request().Context(), req, dataChan, errChan)

	// sse stream response
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	c.Response().WriteHeader(http.StatusOK)

	for {
		select {
		case data := <-dataChan:
			msg, err := json.Marshal(data)
			if err != nil {
				slog.ErrorContext(c.Request().Context(), "completion stream marshal response error", "err", err.Error())
				return c.String(http.StatusInternalServerError, err.Error())
			}
			_, err = c.Response().Write([]byte(fmt.Sprintf("data: %s\n\n", msg)))
			if err != nil {
				slog.ErrorContext(c.Request().Context(), "write completion stream response error", "err", err.Error())
				return c.String(http.StatusInternalServerError, err.Error())
			}
			c.Response().Flush()
		case err := <-errChan:
			if errors.Is(err, io.EOF) {
				return c.String(http.StatusOK, "data: [DONE]\n\n")
			}
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
}
