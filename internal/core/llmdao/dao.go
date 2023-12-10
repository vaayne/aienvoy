package llmdao

import (
	"context"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/internal/pkg/dtoutils"
	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/google/uuid"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
)

type Dao struct {
	tx *daos.Dao
}

func New(tx *daos.Dao) *Dao {
	return &Dao{tx: tx}
}

func (d *Dao) SaveConversation(ctx context.Context, conversation llm.Conversation) (llm.Conversation, error) {
	var cov ConversationDTO
	cov.FromLLMConversation(conversation)

	if cov.UserId == "" {
		cov.UserId = ctx.Value(config.ContextKeyUserId).(string)
	}

	if cov.Id == "" {
		cov.Id = uuid.NewString()
	}

	if err := d.tx.DB().Model(&cov).Insert(); err != nil {
		return llm.Conversation{}, err
	}

	return d.GetConversation(ctx, cov.Id)
}

func (d *Dao) GetConversation(ctx context.Context, id string) (llm.Conversation, error) {
	var dto ConversationDTO
	if err := d.tx.DB().Select().Model(id, &dto); err != nil {
		return llm.Conversation{}, err
	}
	return dto.ToLLMConversation(), nil
}

func (d *Dao) ListConversations(ctx context.Context) ([]llm.Conversation, error) {
	var dtos []ConversationDTO
	if err := d.tx.DB().Select().All(&dtos); err != nil {
		return nil, err
	}

	var conversations []llm.Conversation
	for _, dto := range dtos {
		conversations = append(conversations, dto.ToLLMConversation())
	}

	return conversations, nil
}

func (d *Dao) DeleteConversation(ctx context.Context, id string) error {
	return d.tx.DB().Model(&ConversationDTO{BaseModel: dtoutils.BaseModel{Id: id}}).Delete()
}

func (d *Dao) SaveMessage(ctx context.Context, message llm.Message) (llm.Message, error) {
	var msg MessageDTO
	msg.FromLLMMessage(message)
	if msg.UserId == "" {
		msg.UserId = ctx.Value(config.ContextKeyUserId).(string)
	}
	if msg.Id == "" {
		msg.Id = uuid.NewString()
	}

	if err := d.tx.DB().Model(&msg).Insert(); err != nil {
		return llm.Message{}, err
	}

	return d.GetMessage(ctx, msg.Id)
}

func (d *Dao) GetMessage(ctx context.Context, id string) (llm.Message, error) {
	var dto MessageDTO
	if err := d.tx.DB().Select().Model(id, &dto); err != nil {
		return llm.Message{}, err
	}
	return dto.ToLLMMessage(), nil
}

func (d *Dao) ListMessages(ctx context.Context, conversationId string) ([]llm.Message, error) {
	var dtos []MessageDTO
	if err := d.tx.DB().Select().Where(dbx.HashExp{"conversation_id": conversationId}).All(&dtos); err != nil {
		return nil, err
	}

	var messages []llm.Message
	for _, dto := range dtos {
		messages = append(messages, dto.ToLLMMessage())
	}

	return messages, nil
}

func (d *Dao) DeleteMessage(ctx context.Context, id string) error {
	return d.tx.DB().Model(&MessageDTO{BaseModel: dtoutils.BaseModel{Id: id}}).Delete()
}

func (d *Dao) GetConversationLastMessage(ctx context.Context, id string) (llm.Message, error) {
	var dto MessageDTO
	if err := d.tx.DB().Select().Where(dbx.HashExp{"conversation_id": id}).OrderBy("created DESC").Limit(1).One(&dto); err != nil {
		return llm.Message{}, err
	}
	return dto.ToLLMMessage(), nil
}
