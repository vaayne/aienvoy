package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	innerllm "github.com/Vaayne/aienvoy/internal/core/llm"
	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/pocketbase/pocketbase/daos"

	"github.com/labstack/echo/v5"
)

type LLMHandler struct{}

func NewLLMHandler() *LLMHandler {
	return &LLMHandler{}
}

//func (l *LLMHandler) GetModels(c echo.Context) error {
//	resp, err := llmopenai.New().GetModels(c.Request().Context())
//	if err != nil {
//		slog.ErrorContext(c.Request().Context(), "get models error", "err", err.Error())
//		return c.String(http.StatusInternalServerError, err.Error())
//	}
//	return c.JSON(http.StatusOK, resp)
//}

func (l *LLMHandler) CreateChatCompletion(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(llm.ChatCompletionRequest)
	err := c.Bind(req)
	if err != nil {
		slog.ErrorContext(ctx, "bind chat request body error", "err", err.Error())
		return c.String(http.StatusBadRequest, "bad request")
	}

	svc := innerllm.New(req.Model, innerllm.NewDao(c.Get(config.ContextKeyDao).(*daos.Dao)))
	if svc == nil {
		return c.String(http.StatusBadRequest, "unknown model")
	}

	if req.Stream {
		return l.chatStream(c, svc, *req)
	}

	resp, err := svc.CreateChatCompletion(ctx, *req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (l *LLMHandler) chatStream(c echo.Context, svc innerllm.Service, req llm.ChatCompletionRequest) error {
	dataChan := make(chan llm.ChatCompletionStreamResponse)
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
