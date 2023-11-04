package llm

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

type Dao interface {
	SaveConversation(ctx context.Context, conversation Conversation) (Conversation, error)
	GetConversation(ctx context.Context, id string) (Conversation, error)
	ListConversations(ctx context.Context) ([]Conversation, error)
	DeleteConversation(ctx context.Context, id string) error

	SaveMessage(ctx context.Context, message Message) (Message, error)
	GetMessage(ctx context.Context, id string) (Message, error)
	ListMessages(ctx context.Context, conversationId string) ([]Message, error)
	DeleteMessage(ctx context.Context, id string) error

	GetConversationLastMessage(ctx context.Context, id string) (Message, error)
}

const (
	conversationCachePrefix = "conversation:"
	messageCachePrefix      = "message:"
)

func conversationCacheKey(id string) string {
	return conversationCachePrefix + id
}

func messageCacheKey(id string) string {
	return messageCachePrefix + id
}

var DefaultDao = NewMemoryDao()

type MemoryDao struct {
	client *cache.Cache
}

func NewMemoryDao() *MemoryDao {
	return &MemoryDao{
		client: cache.New(24*time.Hour, 7*24*time.Hour),
	}
}

func (d *MemoryDao) SaveConversation(ctx context.Context, conversation Conversation) (Conversation, error) {
	if conversation.Id == "" {
		conversation.Id = uuid.NewString()
	}
	conversation.CreatedAt = time.Now()

	conversation.UpdatedAt = time.Now()

	d.client.Set(conversationCacheKey(conversation.Id), conversation, cache.DefaultExpiration)
	return conversation, nil
}

func (d *MemoryDao) GetConversation(ctx context.Context, id string) (Conversation, error) {
	conversation, ok := d.client.Get(conversationCacheKey(id))
	if !ok {
		return Conversation{}, errors.New("conversation not found")
	}
	return conversation.(Conversation), nil
}

func (d *MemoryDao) ListConversations(ctx context.Context) ([]Conversation, error) {
	var conversations []Conversation
	for key, val := range d.client.Items() {
		if strings.HasPrefix(key, conversationCachePrefix) {
			conversations = append(conversations, val.Object.(Conversation))
		}
	}
	return conversations, nil
}

func (d *MemoryDao) DeleteConversation(ctx context.Context, id string) error {
	d.client.Delete(conversationCacheKey(id))
	return nil
}

func (d *MemoryDao) SaveMessage(ctx context.Context, message Message) (Message, error) {
	if message.Id == "" {
		message.Id = uuid.NewString()
	}
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	message.Model = message.Request.Model
	d.client.Set(messageCacheKey(message.Id), message, cache.DefaultExpiration)
	return message, nil
}

func (d *MemoryDao) GetMessage(ctx context.Context, id string) (Message, error) {
	message, ok := d.client.Get(messageCacheKey(id))
	if !ok {
		return Message{}, errors.New("message not found")
	}
	return message.(Message), nil
}

func (d *MemoryDao) ListMessages(ctx context.Context, conversationId string) ([]Message, error) {
	var messages []Message
	for key, val := range d.client.Items() {
		if strings.HasPrefix(key, messageCachePrefix) {
			message := val.Object.(Message)
			if message.ConversationId == conversationId {
				messages = append(messages, message)
			}
		}
	}
	return messages, nil
}

func (d *MemoryDao) DeleteMessage(ctx context.Context, id string) error {
	d.client.Delete(messageCacheKey(id))
	return nil
}

func (d *MemoryDao) GetConversationLastMessage(ctx context.Context, id string) (Message, error) {
	messages, err := d.ListMessages(ctx, id)
	if err != nil {
		return Message{}, err
	}
	if len(messages) == 0 {
		return Message{}, errors.New("conversation not found")
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})

	return messages[0], nil
}
