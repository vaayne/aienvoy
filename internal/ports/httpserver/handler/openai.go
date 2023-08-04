package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"aienvoy/internal/core/llm/openai"
	"aienvoy/internal/pkg/context"
	"aienvoy/internal/pkg/logger"

	"github.com/labstack/echo/v5"
)

type OpenAIHandler struct {
	openai *openai.OpenAI
}

func NewOpenAIHandler() *OpenAIHandler {
	return &OpenAIHandler{
		openai: &openai.OpenAI{},
	}
}

func (h *OpenAIHandler) GetModels(c echo.Context) error {
	resp, err := h.openai.GetModels(context.FromEchoContext(c))
	if err != nil {
		logger.GetSugaredLoggerWithEchoContext(c).Errorw("get models error", "err", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *OpenAIHandler) Chat(c echo.Context) error {
	req := new(openai.ChatCompletionRequest)
	err := c.Bind(req)
	if err != nil {
		logger.GetSugaredLoggerWithEchoContext(c).Errorw("bind chat request body error", "err", err.Error())
		return c.String(http.StatusBadRequest, "bad request")
	}
	// stream response
	if req.Stream {
		return h.chatStream(c, req)
	}

	return h.chat(c, req)
}

func (h *OpenAIHandler) CreateEmbeddings(c echo.Context) error {
	var req *openai.EmbeddingRequest
	err := c.Bind(req)
	if err != nil {
		logger.GetSugaredLoggerWithEchoContext(c).Errorw("bind embedding request body error", "err", err.Error())
		return c.String(http.StatusBadRequest, "bad request")
	}

	resp, err := h.openai.CreateEmbeddings(c.Request().Context(), req)
	if err != nil {
		logger.GetSugaredLoggerWithEchoContext(c).Errorw("create embedding error", "err", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *OpenAIHandler) chat(c echo.Context, req *openai.ChatCompletionRequest) error {
	resp, err := h.openai.Chat(context.FromEchoContext(c), req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *OpenAIHandler) chatStream(c echo.Context, req *openai.ChatCompletionRequest) error {
	dataChan := make(chan openai.ChatCompletionStreamResponse)
	defer close(dataChan)
	errChan := make(chan error)
	defer close(errChan)

	go h.openai.ChatStream(context.FromEchoContext(c), req, dataChan, errChan)

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
				logger.GetSugaredLoggerWithEchoContext(c).Errorw("chat stream marshal response error", "err", err.Error())
				return c.String(http.StatusInternalServerError, err.Error())
			}
			_, err = c.Response().Write([]byte(fmt.Sprintf("data: %s\n\n", msg)))
			if err != nil {
				logger.GetSugaredLoggerWithEchoContext(c).Errorw("write chat stream response error", "err", err.Error())
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
