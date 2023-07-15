package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"aienvoy/internal/core/llm/openai"
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
	resp, err := h.openai.GetModels(c.Request().Context())
	if err != nil {
		logger.SugaredLogger.Errorw("get models error", "err", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *OpenAIHandler) Chat(c echo.Context) error {
	req := new(openai.ChatCompletionRequest)
	err := c.Bind(req)
	if err != nil {
		logger.SugaredLogger.Errorw("bind chat request body error", "err", err.Error())
		return c.String(http.StatusBadRequest, "bad request")
	}
	logger.SugaredLogger.Debugw("chat req", "req", req)

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
		logger.SugaredLogger.Infow("bind embedding request body error", "err", err.Error())
		return c.String(http.StatusBadRequest, "bad request")
	}

	resp, err := h.openai.CreateEmbeddings(c.Request().Context(), req)
	if err != nil {
		logger.SugaredLogger.Errorw("create embedding error", "err", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *OpenAIHandler) chat(c echo.Context, req *openai.ChatCompletionRequest) error {
	resp, err := h.openai.Chat(c.Request().Context(), req)
	if err != nil {
		logger.SugaredLogger.Errorw("Chat with OpenAI error", "err", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *OpenAIHandler) chatStream(c echo.Context, req *openai.ChatCompletionRequest) error {
	stream, err := h.openai.ChatStream(c.Request().Context(), req)
	if err != nil {
		logger.SugaredLogger.Errorw("Chat with OpenAI error", "err", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return c.String(http.StatusOK, "data: [DONE]\n\n")
		}
		if msg, err := json.Marshal(resp); err != nil {
			logger.SugaredLogger.Errorw("marshal response error", "err", err.Error())
			return c.String(http.StatusInternalServerError, err.Error())
		} else {
			if _, err = c.Response().Write([]byte(fmt.Sprintf("data: %s\n\n", msg))); err != nil {
				logger.SugaredLogger.Errorw("write response error", "err", err.Error())
				return c.String(http.StatusInternalServerError, err.Error())
			}
		}
		c.Response().Flush()
	}
}
