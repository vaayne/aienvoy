package bard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/Vaayne/aienvoy/pkg/llm"
)

const ModelBard = "bard"

type Bard struct {
	client *Client
	dao    llm.Dao
	*llm.LLM
}

func New(token string, dao llm.Dao, opts ...ClientOption) (*Bard, error) {
	client, err := NewClient(token, opts...)
	if err != nil {
		return nil, err
	}
	return &Bard{
		client: client,
		dao:    dao,
		LLM:    llm.New(dao, client),
	}, nil
}

func (c *Client) ListModels() []string {
	return []string{ModelBard}
}

func (c *Client) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	slog.InfoContext(ctx, "chat with Google Bard start")
	prompt := req.ToPrompt()
	resp, err := c.Ask(prompt, "", "", "", 0)
	if err != nil {
		slog.ErrorContext(ctx, "chat with Google Bard error", "err", err)
		return llm.ChatCompletionResponse{}, fmt.Errorf("bard got an error, %w", err)
	}
	res := resp.ToChatCompletionResponse()
	slog.InfoContext(ctx, "chat with Google Bard success")
	return res, nil
}

func (c *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat with Google Bard stream start")
	prompt := req.ToPrompt()
	resp, err := c.Ask(prompt, "", "", "", 0)
	if err != nil {
		errChan <- fmt.Errorf("bard got an error, %w", err)
		slog.ErrorContext(ctx, "chat with Google Bard stream error", "err", err)
		return
	}
	res := resp.ToChatCompletionStreamResponse()
	slog.InfoContext(ctx, "chat with Google Bard stream success")
	dataChan <- res
	errChan <- io.EOF
}

func (b *Bard) CreateConversation(ctx context.Context, name string) (llm.Conversation, error) {
	prompt := "hello"
	resp, err := b.client.FirstAsk(prompt)
	if err != nil {
		return llm.Conversation{}, fmt.Errorf("bard create conversation error, %w", err)
	}

	req := &llm.ChatCompletionRequest{}
	req.FromPrompt(ModelBard, prompt)

	if _, err := b.saveAnswer(ctx, resp.ConversationID, *req, resp); err != nil {
		return llm.Conversation{}, err
	}
	cov := llm.Conversation{
		Id:        resp.ConversationID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Deleted:   false,
		Name:      name,
		Model:     ModelBard,
	}
	return b.dao.SaveConversation(ctx, cov)
}

func (b *Bard) CreateMessage(ctx context.Context, conversationId string, req llm.ChatCompletionRequest) (llm.Message, error) {
	if conversationId == "" {
		slog.ErrorContext(ctx, "conversation id is empty")
		return llm.Message{}, errors.New("conversation id is empty")
	}
	_, err := b.dao.GetConversation(ctx, conversationId)
	if err != nil {
		slog.ErrorContext(ctx, "get conversation for create message error", "err", err, "conversation_id", conversationId)
		return llm.Message{}, err
	}
	lastMessage, err := b.dao.GetConversationLastMessage(ctx, conversationId)
	if err != nil {
		slog.ErrorContext(ctx, "get conversation last message for create message error", "err", err, "conversation_id", conversationId)
		return llm.Message{}, err
	}

	var lastAnswer Answer
	if err := json.Unmarshal(lastMessage.RawResponse, &lastAnswer); err != nil {
		slog.ErrorContext(ctx, "unmarshal last message raw response error", "err", err, "conversation_id", conversationId)
		return llm.Message{}, err
	}
	prompt := req.ToPromptWithoutRole()

	answer, err := b.client.Ask(prompt, lastAnswer.ConversationID, lastAnswer.ResponseID, lastAnswer.Choices[0].ID, 0)
	if err != nil {
		slog.ErrorContext(ctx, "create message error", "err", err, "conversation_id", conversationId, "model", req.Model)
		return llm.Message{}, err
	}
	message, err := b.saveAnswer(ctx, conversationId, req, answer)
	slog.InfoContext(ctx, "create message", "message", message, "err", err, "model", req.Model)
	return message, err
}

func (b *Bard) CreateMessageStream(ctx context.Context, conversationId string, req llm.ChatCompletionRequest, respChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	if conversationId == "" {
		slog.ErrorContext(ctx, "conversation id is empty")
		errChan <- errors.New("conversation id is empty")
		return
	}
	_, err := b.dao.GetConversation(ctx, conversationId)
	if err != nil {
		slog.ErrorContext(ctx, "get conversation for create message error", "err", err, "conversation_id", conversationId)
		errChan <- fmt.Errorf("bard create message stream error, %w", err)
		return
	}
	lastMessage, err := b.dao.GetConversationLastMessage(ctx, conversationId)
	if err != nil {
		slog.ErrorContext(ctx, "get conversation last message for create message error", "err", err, "conversation_id", conversationId)
		errChan <- fmt.Errorf("bard create message stream error, %w", err)
		return
	}

	var lastAnswer Answer
	if err := json.Unmarshal(lastMessage.RawResponse, &lastAnswer); err != nil {
		slog.ErrorContext(ctx, "unmarshal last message raw response error", "err", err, "conversation_id", conversationId)
		errChan <- fmt.Errorf("bard create message stream error, %w", err)
		return
	}
	prompt := req.ToPromptWithoutRole()

	answer, err := b.client.Ask(prompt, lastAnswer.ConversationID, lastAnswer.ResponseID, lastAnswer.Choices[0].ID, 0)
	if err != nil {
		slog.ErrorContext(ctx, "create message error", "err", err, "conversation_id", conversationId, "model", req.Model)
		errChan <- fmt.Errorf("bard create message stream error, %w", err)
		return
	}

	respChan <- answer.ToChatCompletionStreamResponse()

	if _, err := b.saveAnswer(ctx, conversationId, req, answer); err != nil {
		slog.ErrorContext(ctx, "save answer error", "err", err, "conversation_id", conversationId, "model", req.Model)
	}

	slog.InfoContext(ctx, "create message success", "model", req.Model)
	errChan <- io.EOF
}

func (b *Bard) saveAnswer(ctx context.Context, conversationId string, req llm.ChatCompletionRequest, answer *Answer) (llm.Message, error) {
	res := answer.ToChatCompletionResponse()
	rawResp, _ := json.Marshal(answer)
	message, err := b.dao.SaveMessage(ctx, llm.Message{
		Id:             answer.ResponseID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ConversationId: conversationId,
		Model:          req.Model,
		Request:        req,
		Response:       res,
		RawResponse:    rawResp,
	})
	if err != nil {
		err = fmt.Errorf("bard save answer error, %w", err)
	}
	return message, err
}
