package llm

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
)

var NotImplementError = errors.New("the method not implement")

type client interface {
	ListModels() []string
	CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error)
	CreateChatCompletionStream(ctx context.Context, req ChatCompletionRequest, respChan chan ChatCompletionStreamResponse, errChan chan error)
}

type LLM struct {
	client
	dao Dao
}

func New(dao Dao, c client) *LLM {
	return &LLM{
		dao:    dao,
		client: c,
	}
}

func (l *LLM) CreateConversation(ctx context.Context, name string) (Conversation, error) {
	cov, err := l.dao.SaveConversation(ctx, Conversation{
		Id:        uuid.NewString(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	slog.InfoContext(ctx, "create conversation", "cov", cov, "err", err)
	return cov, err
}

func (l *LLM) ListConversations(ctx context.Context) ([]Conversation, error) {
	conversations, err := l.dao.ListConversations(ctx)
	slog.InfoContext(ctx, "list conversations", "conversations", conversations, "err", err)
	return conversations, err
}

func (l *LLM) GetConversation(ctx context.Context, id string) (Conversation, error) {
	cov, err := l.dao.GetConversation(ctx, id)
	slog.InfoContext(ctx, "get conversation", "cov", cov, "err", err)
	return cov, err
}

func (l *LLM) DeleteConversation(ctx context.Context, id string) error {
	err := l.dao.DeleteConversation(ctx, id)
	slog.Info("delete conversation", "id", id, "err", err)
	return err
}

func (l *LLM) CreateMessage(ctx context.Context, conversationId string, req ChatCompletionRequest) (Message, error) {
	if conversationId == "" {
		slog.ErrorContext(ctx, "conversation id is empty")
		return Message{}, errors.New("conversation id is empty")
	}
	cov, err := l.dao.GetConversation(ctx, conversationId)
	if err != nil {
		slog.ErrorContext(ctx, "get conversation for create message error", "err", err, "conversation_id", conversationId)
		return Message{}, err
	}

	messages, err := l.ListMessages(ctx, cov.Id)
	if err != nil {
		slog.ErrorContext(ctx, "list messages for create message error", "err", err, "conversation_id", conversationId)
		return Message{}, err
	}
	originReqMessages := req.Messages
	reqMessages := make([]ChatCompletionMessage, 0)
	// add history message to request
	for _, message := range messages {
		// latest request message as prompt
		reqMessages = append(reqMessages, message.Request.Messages...)
		// add response message as prompt
		reqMessages = append(reqMessages, message.Response.Choices[0].Message)
	}
	reqMessages = append(reqMessages, originReqMessages...)
	req.Messages = reqMessages
	resp, err := l.client.CreateChatCompletion(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "create message error", "err", err, "conversation_id", conversationId, "model", req.Model)
		return Message{}, err
	}
	req.Messages = originReqMessages
	message, err := l.dao.SaveMessage(ctx, Message{
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ConversationId: conversationId,
		Model:          req.Model,
		Request:        req,
		Response:       resp,
	})
	slog.InfoContext(ctx, "create message", "message", message, "err", err)
	return message, err
}

func (l *LLM) CreateMessageStream(ctx context.Context, conversationId string, req ChatCompletionRequest, respChan chan ChatCompletionStreamResponse, errChan chan error) {
	if conversationId == "" {
		slog.ErrorContext(ctx, "conversation id is empty")
		errChan <- errors.New("conversation id is empty")
		return
	}
	_, err := l.dao.GetConversation(ctx, conversationId)
	if err != nil {
		slog.ErrorContext(ctx, "get conversation for create message error", "err", err, "conversation_id", conversationId)
		errChan <- err
		return
	}

	messages, err := l.ListMessages(ctx, conversationId)
	if err != nil {
		slog.ErrorContext(ctx, "list messages for create message error", "err", err, "conversation_id", conversationId)
		errChan <- err
		return
	}
	originReqMessages := req.Messages
	reqMessages := make([]ChatCompletionMessage, 0)
	// add history message to request
	for _, message := range messages {
		// latest request message as prompt
		reqMessages = append(reqMessages, message.Request.Messages...)
		// add response message as prompt
		reqMessages = append(reqMessages, message.Response.Choices[0].Message)
	}
	reqMessages = append(reqMessages, originReqMessages...)
	req.Messages = reqMessages

	slog.InfoContext(ctx, "create message stream", "req", req)

	innerDataChan := make(chan ChatCompletionStreamResponse)
	innerErrChan := make(chan error)

	go l.client.CreateChatCompletionStream(ctx, req, innerDataChan, innerErrChan)

	sb := strings.Builder{}

	var resp ChatCompletionStreamResponse

	for {
		select {
		case resp = <-innerDataChan:
			sb.WriteString(resp.Choices[0].Delta.Content)
			respChan <- resp
		case err := <-innerErrChan:
			if errors.Is(err, io.EOF) {
				req.Messages = originReqMessages
				chatCompletionResponse := resp.ToChatCompletionResponse()
				chatCompletionResponse.Choices[0].Message.Content = sb.String()
				if _, err := l.dao.SaveMessage(ctx, Message{
					Id:             resp.ID,
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
					ConversationId: conversationId,
					Model:          req.Model,
					Request:        req,
					Response:       chatCompletionResponse,
				}); err != nil {
					slog.ErrorContext(ctx, "save message error", "err", err)
				}
			} else {
				slog.ErrorContext(ctx, "create message error", "err", err, "conversation_id", conversationId)
			}
			errChan <- err
			return
		case <-ctx.Done():
			errChan <- ctx.Err()
		}
	}
}

func (l *LLM) ListMessages(ctx context.Context, conversationId string) ([]Message, error) {
	messages, err := l.dao.ListMessages(ctx, conversationId)
	slog.InfoContext(ctx, "list messages", "conversation_id", conversationId, "err", err)
	return messages, err
}

func (l *LLM) GetMessage(ctx context.Context, id string) (Message, error) {
	message, err := l.dao.GetMessage(ctx, id)
	slog.InfoContext(ctx, "get message", "message_id", message, "err", err)
	return message, err
}

func (l *LLM) DeleteMessage(ctx context.Context, id string) error {
	err := l.dao.DeleteMessage(ctx, id)
	slog.InfoContext(ctx, "delete message", "id", id, "err", err)
	return err
}
