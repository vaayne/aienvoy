package llmservice

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Vaayne/aienvoy/internal/pkg/dtoutils"
	"github.com/Vaayne/aienvoy/pkg/llm"
	"github.com/pocketbase/pocketbase/tools/types"
)

const (
	tableNameConversations = "conversations"
	tableNameMessages      = "conversation_messages"
)

type ConversationDTO struct {
	dtoutils.BaseModel
	UserId    string `json:"user_id"  db:"user_id"`
	Name      string `json:"name,omitempty"  db:"name"`
	Model     string `json:"model,omitempty" db:"model"`
	Summary   string `json:"summary,omitempty" db:"summary"`
	ExtraInfo string `json:"extra_info,omitempty" db:"extra_info"`
}

func (c ConversationDTO) TableName() string {
	return tableNameConversations
}

func (c *ConversationDTO) FromLLMConversation(conversation llm.Conversation) {
	c.Id = conversation.Id
	c.Created = mustParseDateTime(conversation.CreatedAt)
	c.Updated = mustParseDateTime(conversation.UpdatedAt)
	c.UserId = conversation.UserId
	c.Name = conversation.Name
	c.Model = conversation.Model
	c.Summary = conversation.Summary
	c.ExtraInfo = conversation.ExtraInfo
}

func (c ConversationDTO) ToLLMConversation() llm.Conversation {
	return llm.Conversation{
		Id:        c.Id,
		CreatedAt: c.Created.Time(),
		UpdatedAt: c.Updated.Time(),
		UserId:    c.UserId,
		Name:      c.Name,
		Model:     c.Model,
		Summary:   c.Summary,
		ExtraInfo: c.ExtraInfo,
	}
}

type MessageDTO struct {
	dtoutils.BaseModel
	UserId          string `json:"user_id"  db:"user_id"`
	ConversationId  string `json:"conversation_id"  db:"conversation_id"`
	Model           string `json:"model,omitempty" db:"model"`
	PromptToken     string `json:"prompt_token,omitempty" db:"prompt_token"`
	CompletionToken string `json:"completion_token,omitempty" db:"completion_token"`
	Description     string `json:"description,omitempty" db:"description"`
	Request         []byte `json:"request,omitempty" db:"request"`
	Response        []byte `json:"response,omitempty" db:"response"`
	RawResponse     []byte `json:"raw_response,omitempty" db:"raw_response"`
}

func (m MessageDTO) TableName() string {
	return tableNameMessages
}

func (m *MessageDTO) FromLLMMessage(message llm.Message) {
	m.Id = message.Id
	m.Created = mustParseDateTime(message.CreatedAt)
	m.Updated = mustParseDateTime(message.UpdatedAt)
	m.UserId = message.UserId
	m.ConversationId = message.ConversationId
	m.Model = message.Model
	// m.PromptToken = message.PromptToken
	// m.CompletionToken = message.CompletionToken
	m.Description = message.Description

	m.Request = mustMarshal(message.Request)
	m.Response = mustMarshal(message.Response)
	m.RawResponse = message.RawResponse
}

func (m MessageDTO) ToLLMMessage() llm.Message {
	var req llm.ChatCompletionRequest
	var resp llm.ChatCompletionResponse
	mustUnMarshal(m.Request, &req)
	mustUnMarshal(m.Response, &resp)
	return llm.Message{
		Id:             m.Id,
		CreatedAt:      m.Created.Time(),
		UpdatedAt:      m.Updated.Time(),
		UserId:         m.UserId,
		ConversationId: m.ConversationId,
		Model:          m.Model,
		// PromptToken:     m.PromptToken,
		// CompletionToken: m.CompletionToken,
		Description: m.Description,
		Request:     req,
		Response:    resp,
		RawResponse: m.RawResponse,
	}
}

func mustParseDateTime(t time.Time) types.DateTime {
	dt, err := types.ParseDateTime(t)
	if err != nil {
		panic(err)
	}
	return dt
}

func mustMarshal(in any) []byte {
	out, err := json.Marshal(in)
	if err != nil {
		slog.Error("must marshal object error", "err", err)
	}
	return out
}

// mustUnMarshal
// out must be pointer
func mustUnMarshal(in []byte, out any) {
	if err := json.Unmarshal(in, out); err != nil {
		slog.Error("must unmarshal object error", "err", err)
	}
}
