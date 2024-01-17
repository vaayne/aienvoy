package claudeweb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/Vaayne/aienvoy/pkg/llms/llm"
)

// ClaudeWeb is a Claude request client
type ClaudeWeb struct {
	client *Client
	dao    llm.Dao
	*llm.LLM
}

func New(sessionKey string, dao llm.Dao) *ClaudeWeb {
	client := NewClient(sessionKey)
	return &ClaudeWeb{
		client: client,
		dao:    dao,
		LLM:    llm.New(dao, client),
	}
}

func (cw *Client) ListModels() []string {
	return ListModels()
}

func ListModels() []string {
	return []string{ModelClaude2, ModelClaude2Dot1}
}

func (cw *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat with Claude Web stream start")
	prompt := req.ToPrompt()
	cov, err := cw.CreateConversation(prompt[:min(10, len(prompt))])
	if err != nil {
		errChan <- fmt.Errorf("create new claude conversiton error: %w", err)
		return
	}

	messageChan := make(chan *ChatMessageResponse)
	innerErrChan := make(chan error)

	go cw.CreateChatMessageStream(cov.UUID, prompt, messageChan, innerErrChan)
	sb := strings.Builder{}
	for {
		select {
		case resp := <-messageChan:
			sb.WriteString(resp.Completion)
			dataChan <- resp.ToChatCompletionStreamResponse()
		case err := <-innerErrChan:
			if errors.Is(err, io.EOF) {
				slog.InfoContext(ctx, "claude stream done", "cov_id", cov.UUID)
				slog.InfoContext(ctx, "chat with Claude Web stream success")
				errChan <- err
				return
			}
			slog.ErrorContext(ctx, "chat with Claude Web error", "err", err)
			errChan <- err
			return
		}
	}
}

func (c *ClaudeWeb) CreateConversation(ctx context.Context, name string) (llm.Conversation, error) {
	cov, err := c.client.CreateConversation(name)
	if err != nil {
		return llm.Conversation{}, fmt.Errorf("create new claude conversiton error: %w", err)
	}
	return c.dao.SaveConversation(ctx, cov.ToLLMConversation())
}

func (c *ClaudeWeb) CreateMessage(ctx context.Context, conversationId string, req llm.ChatCompletionRequest) (llm.Message, error) {
	if conversationId == "" {
		slog.ErrorContext(ctx, "conversation id is empty")
		return llm.Message{}, errors.New("conversation id is empty")
	}
	_, err := c.dao.GetConversation(ctx, conversationId)
	if err != nil {
		slog.ErrorContext(ctx, "get conversation for create message error", "err", err, "conversation_id", conversationId)
		return llm.Message{}, err
	}

	resp, err := c.client.CreateChatMessage(conversationId, req.ToPromptWithoutRole())
	if err != nil {
		slog.ErrorContext(ctx, "chat with Claude Web error", "err", err)
		return llm.Message{}, fmt.Errorf("chat with claude error: %v", err)
	}
	message, err := c.saveResponseMessage(ctx, conversationId, req, *resp)
	slog.InfoContext(ctx, "chat with Claude Web success")
	return message, err
}

func (c *ClaudeWeb) CreateMessageStream(ctx context.Context, conversationId string, req llm.ChatCompletionRequest, respChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	if conversationId == "" {
		slog.ErrorContext(ctx, "conversation id is empty")
		errChan <- errors.New("conversation id is empty")
		return
	}
	_, err := c.dao.GetConversation(ctx, conversationId)
	if err != nil {
		slog.ErrorContext(ctx, "get conversation for create message error", "err", err, "conversation_id", conversationId)
		errChan <- fmt.Errorf("bard create message stream error, %w", err)
		return
	}

	messageChan := make(chan *ChatMessageResponse)
	innerErrChan := make(chan error)
	var resp *ChatMessageResponse

	go c.client.CreateChatMessageStream(conversationId, req.ToPromptWithoutRole(), messageChan, innerErrChan)
	sb := strings.Builder{}
	for {
		select {
		case resp = <-messageChan:
			sb.WriteString(resp.Completion)
			respChan <- resp.ToChatCompletionStreamResponse()
		case err := <-innerErrChan:
			if errors.Is(err, io.EOF) {
				slog.InfoContext(ctx, "claude stream done", "cov_id", conversationId)
				resp.Completion = sb.String()
				_, _ = c.saveResponseMessage(ctx, conversationId, req, *resp)
				errChan <- io.EOF
				return
			}
			slog.ErrorContext(ctx, "chat with Claude Web error", "err", err)
			errChan <- err
			return
		}
	}
}

func (c *ClaudeWeb) saveResponseMessage(ctx context.Context, conversationId string, req llm.ChatCompletionRequest, resp ChatMessageResponse) (llm.Message, error) {
	res := resp.ToChatCompletionResponse()
	rawResp, _ := json.Marshal(res)
	message, err := c.dao.SaveMessage(ctx, llm.Message{
		Id:             resp.LogId,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ConversationId: conversationId,
		Model:          req.ModelId(),
		Request:        req,
		Response:       res,
		RawResponse:    rawResp,
	})
	if err != nil {
		err = fmt.Errorf("claudeweb save answer error, %w", err)
	}
	return message, err
}
