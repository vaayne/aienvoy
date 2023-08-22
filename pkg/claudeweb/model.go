package claudeweb

type Organization struct {
	UUID         string   `json:"uuid"`
	Name         string   `json:"name"`
	Settings     Settings `json:"settings"`
	Capabilities []string `json:"capabilities"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	ActiveFlags  []string `json:"active_flags"`
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
