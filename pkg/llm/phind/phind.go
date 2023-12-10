package phind

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llm"
)

type Phind struct {
	*llm.LLM
}

func New(cookies []*http.Cookie, dao llm.Dao) *Phind {
	return &Phind{
		LLM: llm.New(dao, NewClient(cookies)),
	}
}

const ModelPhindV1 = "phind"

func (p *Client) ListModels() []string {
	return []string{ModelPhindV1}
}

func (p *Client) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	slog.InfoContext(ctx, "chat start", "llm", req.Model, "is_stream", false)
	payload := &Request{}
	payload.FromChatCompletionRequest(req)

	messageChan := make(chan llm.ChatCompletionStreamResponse)
	innerErrChan := make(chan error)

	go p.CreateChatCompletionStream(ctx, req, messageChan, innerErrChan)
	data := llm.ChatCompletionStreamResponse{}
	sb := strings.Builder{}
	for {
		select {
		case data = <-messageChan:
			if data.Choices[0].Delta.Content != "" {
				sb.WriteString(data.Choices[0].Delta.Content)
			}
		case err := <-innerErrChan:
			if errors.Is(err, io.EOF) {
				slog.InfoContext(ctx, "chat success", "llm", req.Model, "is_stream", false)
				return llm.ChatCompletionResponse{
					ID:      data.ID,
					Object:  data.Object,
					Created: data.Created,
					Choices: []llm.ChatCompletionChoice{
						{
							Index: 0,
							Message: llm.ChatCompletionMessage{
								Role:    llm.ChatMessageRoleAssistant,
								Content: sb.String(),
							},
							FinishReason: llm.FinishReason(data.Choices[0].FinishReason),
						},
					},
				}, nil
			}
			slog.ErrorContext(ctx, "chat error", "llm", req.Model, "is_stream", false, "err", err)
			return llm.ChatCompletionResponse{}, err
		case <-ctx.Done():
			return llm.ChatCompletionResponse{}, fmt.Errorf("context done")
		}
	}
}

func (p *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat start", "llm", req.Model, "is_stream", true)
	payload := &Request{}
	payload.FromChatCompletionRequest(req)

	messageChan := make(chan llm.ChatCompletionStreamResponse)
	innerErrChan := make(chan error)

	go p.CreateCompletion(ctx, payload, messageChan, innerErrChan)
	sb := strings.Builder{}

	for {
		select {
		case resp := <-messageChan:
			if resp.Choices[0].Delta.Content != "" {
				sb.WriteString(resp.Choices[0].Delta.Content)
			}
			dataChan <- resp
		case err := <-innerErrChan:
			if errors.Is(err, io.EOF) {
				slog.InfoContext(ctx, "chat success", "llm", req.Model, "is_stream", true)
				errChan <- err
				return
			}
			slog.ErrorContext(ctx, "chat error", "llm", req.Model, "is_stream", true, "err", err)
			errChan <- err
			return
		}
	}
}
