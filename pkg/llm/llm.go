package llm

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"time"
)

var NotImplementError = errors.New("the method not implement")

type LLM struct {
	dao Dao
}

func (l *LLM) ListModels() []string {
	panic(NotImplementError)
}

func (l *LLM) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error) {
	panic(NotImplementError)
}

func (l *LLM) CreateChatCompletionStream(ctx context.Context, req ChatCompletionRequest, respChan chan ChatCompletionStreamResponse, errChan chan error) {
	panic(NotImplementError)
}

func (l *LLM) CreateConversation(ctx context.Context, name string) (Conversation, error) {
	cov, err := l.dao.SaveConversation(ctx, Conversation{
		Name: name,
	})
	slog.Info("create conversation", "cov", cov, "err", err)
	return cov, err
}

func (l *LLM) ListConversations(ctx context.Context) ([]Conversation, error) {
	conversations, err := l.dao.ListConversations(ctx)
	slog.Info("list conversations", "conversations", conversations, "err", err)
	return conversations, err
}

func (l *LLM) GetConversation(ctx context.Context, id string) (Conversation, error) {
	cov, err := l.dao.GetConversation(ctx, id)
	slog.Info("get conversation", "cov", cov, "err", err)
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
	resp, err := l.CreateChatCompletion(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "create message error", "err", err, "conversation_id", conversationId)
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
	slog.Info("create message", "message", message, "err", err)
	return message, err
}

func (l *LLM) CreateMessageStream(ctx context.Context, conversationId string, req ChatCompletionRequest, respChan chan ChatCompletionStreamResponse, errChan chan error) {
	if conversationId == "" {
		slog.ErrorContext(ctx, "conversation id is empty")
		errChan <- errors.New("conversation id is empty")
		return
	}
	cov, err := l.dao.GetConversation(ctx, conversationId)
	if err != nil {
		slog.ErrorContext(ctx, "get conversation for create message error", "err", err, "conversation_id", conversationId)
		errChan <- err
		return
	}

	messages, err := l.ListMessages(ctx, cov.Id)
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

	go l.CreateChatCompletionStream(ctx, req, respChan, errChan)

	sb := strings.Builder{}

	var resp ChatCompletionStreamResponse

	for {
		select {
		case resp = <-respChan:
			sb.WriteString(resp.Choices[0].Delta.Content)
		case err := <-errChan:
			if errors.Is(err, io.EOF) {
				req.Messages = originReqMessages
				chatCompletionResponse := resp.ToChatCompletionResponse()
				chatCompletionResponse.Choices[0].Message.Content = sb.String()
				_, _ = l.dao.SaveMessage(ctx, Message{
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
					ConversationId: conversationId,
					Model:          req.Model,
					Request:        req,
					Response:       chatCompletionResponse,
				})
			} else {
				slog.Error("create message error", "err", err, "conversation_id", conversationId)
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
	slog.Info("list messages", "conversation_id", conversationId, "err", err)
	return messages, err
}

func (l *LLM) GetMessage(ctx context.Context, id string) (Message, error) {
	message, err := l.dao.GetMessage(ctx, id)
	slog.Info("get message", "message_id", message, "err", err)
	return message, err
}

func (l *LLM) DeleteMessage(ctx context.Context, id string) error {
	err := l.dao.DeleteMessage(ctx, id)
	slog.Info("delete message", "id", id, "err", err)
	return err
}
