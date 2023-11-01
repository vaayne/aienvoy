package phind

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llmutils"
	"github.com/sashabaranov/go-openai"
)

func (p *Phind) CreateChatCompletion(ctx context.Context, req *openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	slog.InfoContext(ctx, "chat with Phind stream start")
	payload := &Request{}
	payload.FromOpenAIChatCompletionRequest(req)

	messageChan := make(chan openai.ChatCompletionStreamResponse)
	innerErrChan := make(chan error)

	go p.CreateChatCompletionStream(ctx, req, messageChan, innerErrChan)
	data := openai.ChatCompletionStreamResponse{}
	sb := strings.Builder{}
	funcCallArgsBuilder := strings.Builder{}
	for {
		select {
		case data = <-messageChan:
			if data.Choices[0].Delta.Content != "" {
				sb.WriteString(data.Choices[0].Delta.Content)
			}
			call := data.Choices[0].Delta.FunctionCall
			if call != nil {
				funcCallArgsBuilder.WriteString(data.Choices[0].Delta.FunctionCall.Arguments)
			}
		case err := <-innerErrChan:
			if errors.Is(err, io.EOF) {
				return openai.ChatCompletionResponse{
					ID:      data.ID,
					Object:  data.Object,
					Created: data.Created,
					Choices: []openai.ChatCompletionChoice{
						{
							Index: 0,
							Message: openai.ChatCompletionMessage{
								Role:    "assistant",
								Content: sb.String(),
								//FunctionCall: &openai.FunctionCall{Arguments: funcCallArgsBuilder.String(), Name: data.Choices[0].Delta.FunctionCall.Name},
							},
							FinishReason: openai.FinishReason(data.Choices[0].FinishReason),
						},
					},
				}, nil
			}
			return openai.ChatCompletionResponse{}, err
		case <-ctx.Done():
			return openai.ChatCompletionResponse{}, fmt.Errorf("context done")
		}
	}
}

func (p *Phind) CreateChatCompletionStream(ctx context.Context, req *openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat with Phind stream start")
	payload := &Request{}
	payload.FromOpenAIChatCompletionRequest(req)

	messageChan := make(chan openai.ChatCompletionStreamResponse)
	innerErrChan := make(chan error)

	go p.createCompletion(ctx, payload, messageChan, innerErrChan)
	funcCallName := ""
	funcCallArgsSb := strings.Builder{}
	contentSb := strings.Builder{}

	for {
		select {
		case resp := <-messageChan:
			if resp.Choices[0].Delta.Content != "" {
				contentSb.WriteString(resp.Choices[0].Delta.Content)
			}
			if resp.Choices[0].Delta.FunctionCall != nil {
				call := resp.Choices[0].Delta.FunctionCall
				funcCallName = call.Name
				funcCallArgsSb.WriteString(call.Arguments)
			}
			dataChan <- resp
		case err := <-innerErrChan:
			if errors.Is(err, io.EOF) {
				slog.InfoContext(ctx, "chat with Phind stream success", "function call name", funcCallName, "function call args", funcCallArgsSb.String(), "content", contentSb.String())
				errChan <- err
				return
			}
			slog.ErrorContext(ctx, "chat with Phind error", "err", err)
			errChan <- err
			return
		}
	}
}

func (p *Phind) CreateCompletion(ctx context.Context, req *openai.CompletionRequest) (openai.CompletionResponse, error) {
	return llmutils.CreateCompletion(ctx, req, p.CreateChatCompletion)
}

func (p *Phind) CreateCompletionStream(ctx context.Context, req *openai.CompletionRequest, dataChan chan openai.CompletionResponse, errChan chan error) {
	llmutils.CreateCompletionStream(ctx, req, dataChan, errChan, p.CreateChatCompletionStream)
}
