package claudeweb

import (
	"fmt"
	"time"

	"github.com/Vaayne/aienvoy/pkg/llm"
)

type Organization struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Settings struct {
		ClaudeConsolePrivacy string `json:"claude_console_privacy"`
	} `json:"settings"`
	Capabilities             []string    `json:"capabilities"`
	BillableUsagePausedUntil interface{} `json:"billable_usage_paused_until"`
	CreatedAt                time.Time   `json:"created_at"`
	UpdatedAt                time.Time   `json:"updated_at"`
	ActiveFlags              []struct {
		Id          string      `json:"id"`
		Type        string      `json:"type"`
		CreatedAt   time.Time   `json:"created_at"`
		DismissedAt interface{} `json:"dismissed_at"`
		ExpiresAt   interface{} `json:"expires_at"`
	} `json:"active_flags"`
}

type Settings struct {
	ClaudeConsolePrivacy string `json:"claude_console_privacy"`
}

type Conversation struct {
	UUID         string        `json:"uuid"`
	Name         string        `json:"name"`
	Summary      string        `json:"summary"`
	CreatedAt    string        `json:"created_at"`
	UpdatedAt    string        `json:"updated_at"`
	ChatMessages []ChatMessage `json:"chat_messages"`
}

func (c Conversation) ToLLMConversation() llm.Conversation {
	return llm.Conversation{
		Id:      c.UUID,
		Name:    c.Name,
		Summary: c.Summary,
		Model:   ModelClaudeWeb,
	}
}

type Attachment struct {
	FileName         string `json:"file_name"`
	FileType         string `json:"file_type"`
	FileSize         int    `json:"file_size"`
	ExtractedContent string `json:"extracted_content"`
}

type ChatMessage struct {
	UUID         string       `json:"uuid"`
	Text         string       `json:"text"`
	Sender       string       `json:"sender"`
	Index        int          `json:"index"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
	EditedAt     string       `json:"edited_at"`
	ChatFeedback string       `json:"chat_feedback"`
	Attachments  []Attachment `json:"attachments"`
}

type Completion struct {
	Prompt   string `json:"prompt"`
	Timezone string `json:"timezone"`
	Model    string `json:"model"`
}

type CreateChatMessageRequest struct {
	Completion       Completion   `json:"completion"`
	OrganizationUUID string       `json:"organization_uuid"`
	ConversationUUID string       `json:"conversation_uuid"`
	Text             string       `json:"text"`
	Attachments      []Attachment `json:"attachments"`
}

type UpdateConversationRequest struct {
	OrganizationUUID string `json:"organization_uuid"`
	ConversationUUID string `json:"conversation_uuid"`
	Title            string `json:"title"`
}

type ChatMessageResponse struct {
	Completion   string `json:"completion"`
	StopReason   string `json:"stop_reason"`
	Model        string `json:"model"`
	Stop         string `json:"stop"`
	LogId        string `json:"log_id"`
	MessageLimit struct {
		Type string `json:"type"`
	} `json:"message_limit"`
}

func (cr *ChatMessageResponse) stopReasonMapping() string {
	switch cr.StopReason {
	case "stop_sequence":
		return "stop"
	case "max_tokens":
		return "length"
	default:
		return cr.StopReason
	}
}

func (cr *ChatMessageResponse) ToChatCompletionResponse() llm.ChatCompletionResponse {
	return llm.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", cr.LogId),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []llm.ChatCompletionChoice{
			{
				Index: 0,
				Message: llm.ChatCompletionMessage{
					Role:    "assistant",
					Content: cr.Completion,
				},
				FinishReason: llm.FinishReason(cr.stopReasonMapping()),
			},
		},
	}
}

func (cr *ChatMessageResponse) ToChatCompletionStreamResponse() llm.ChatCompletionStreamResponse {
	return llm.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", cr.LogId),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []llm.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: llm.ChatCompletionStreamChoiceDelta{
					Role:    "assistant",
					Content: cr.Completion,
				},
				FinishReason: llm.FinishReason(cr.stopReasonMapping()),
			},
		},
	}
}
