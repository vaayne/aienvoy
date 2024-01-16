package llm

import (
	"fmt"
	"strings"
	"time"
)

// Chat message role defined by the OpenAI API.
const (
	ChatMessageRoleSystem    = "system"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
	ChatMessageRoleFunction  = "function"
)

type Hate struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}
type SelfHarm struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}
type Sexual struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}
type Violence struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}

type ContentFilterResults struct {
	Hate     Hate     `json:"hate,omitempty"`
	SelfHarm SelfHarm `json:"self_harm,omitempty"`
	Sexual   Sexual   `json:"sexual,omitempty"`
	Violence Violence `json:"violence,omitempty"`
}

type PromptAnnotation struct {
	PromptIndex          int                  `json:"prompt_index,omitempty"`
	ContentFilterResults ContentFilterResults `json:"content_filter_results,omitempty"`
}

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`

	// This property isn't in the official documentation, but it's in
	// the documentation for the official library for python:
	// - https://github.com/openai/openai-python/blob/main/chatml.md
	// - https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	Name string `json:"name,omitempty"`

	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

type FunctionCall struct {
	Name string `json:"name,omitempty"`
	// call function with arguments in JSON format
	Arguments string `json:"arguments,omitempty"`
}

// ChatCompletionRequest represents a request structure for chat completion API.
type ChatCompletionRequest struct {
	Model            string                  `json:"model"`
	Messages         []ChatCompletionMessage `json:"messages"`
	MaxTokens        int                     `json:"max_tokens,omitempty"`
	Temperature      float32                 `json:"temperature,omitempty"`
	TopP             float32                 `json:"top_p,omitempty"`
	N                int                     `json:"n,omitempty"`
	Stream           bool                    `json:"stream,omitempty"`
	Stop             []string                `json:"stop,omitempty"`
	PresencePenalty  float32                 `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32                 `json:"frequency_penalty,omitempty"`
	// LogitBias is must be a token id string (specified by their token ID in the tokenizer), not a word string.
	// incorrect: `"logit_bias":{"You": 6}`, correct: `"logit_bias":{"1639": 6}`
	// refs: https://platform.openai.com/docs/api-reference/chat/create#chat/create-logit_bias
	LogitBias    map[string]int       `json:"logit_bias,omitempty"`
	User         string               `json:"user,omitempty"`
	Functions    []FunctionDefinition `json:"functions,omitempty"`
	FunctionCall any                  `json:"function_call,omitempty"`
}

func (r *ChatCompletionRequest) ToPrompt() string {
	return r.toPrompt(true)
}

func (r *ChatCompletionRequest) ToPromptWithoutRole() string {
	return r.toPrompt(false)
}

func (r *ChatCompletionRequest) toPrompt(withRole bool) string {
	sb := strings.Builder{}
	for _, message := range r.Messages {
		if withRole {
			sb.WriteString(fmt.Sprintf("\n\n%s: ", message.Role))
		}
		sb.WriteString(message.Content)
	}

	return sb.String()
}

func (r *ChatCompletionRequest) FromPrompt(model, prompt string) {
	r.Model = model
	r.Messages = []ChatCompletionMessage{
		{
			Role:    ChatMessageRoleUser,
			Content: prompt,
		},
	}
}

func (r *ChatCompletionRequest) ModelId() string {
	texts := strings.Split(r.Model, "/")
	if len(texts) == 1 {
		return r.Model
	}
	return strings.Join(texts[1:], "/")
}

type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Parameters is an object describing the function.
	// You can pass json.RawMessage to describe the schema,
	// or you can pass in a struct which serializes to the proper JSON schema.
	// The jsonschema package is provided for convenience, but you should
	// consider another specialized library if you require more complex schemas.
	Parameters any `json:"parameters"`
}

// Deprecated: use FunctionDefinition instead.
type FunctionDefine = FunctionDefinition

type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonLength        FinishReason = "length"
	FinishReasonFunctionCall  FinishReason = "function_call"
	FinishReasonContentFilter FinishReason = "content_filter"
	FinishReasonNull          FinishReason = "null"
)

func (r FinishReason) MarshalJSON() ([]byte, error) {
	if r == FinishReasonNull || r == "" {
		return []byte("null"), nil
	}
	return []byte(`"` + string(r) + `"`), nil // best effort to not break future API changes
}

type ChatCompletionChoice struct {
	Index   int                   `json:"index"`
	Message ChatCompletionMessage `json:"message"`
	// FinishReason
	// stop: API returned complete message,
	// or a message terminated by one of the stop sequences provided via the stop parameter
	// length: Incomplete model output due to max_tokens parameter or token limit
	// function_call: The model decided to call a function
	// content_filter: Omitted content due to a flag from our content filters
	// null: API response still in progress or incomplete
	FinishReason FinishReason `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse represents a response structure for chat completion API.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage"`
}

type ChatCompletionStreamChoiceDelta struct {
	Content      string        `json:"content,omitempty"`
	Role         string        `json:"role,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

type ChatCompletionStreamChoice struct {
	Index                int                             `json:"index"`
	Delta                ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason         FinishReason                    `json:"finish_reason"`
	ContentFilterResults ContentFilterResults            `json:"content_filter_results,omitempty"`
}

type ChatCompletionStreamResponse struct {
	ID                string                       `json:"id"`
	Object            string                       `json:"object"`
	Created           int64                        `json:"created"`
	Model             string                       `json:"model"`
	Choices           []ChatCompletionStreamChoice `json:"choices"`
	PromptAnnotations []PromptAnnotation           `json:"prompt_annotations,omitempty"`
}

func (r *ChatCompletionStreamResponse) ToChatCompletionResponse() ChatCompletionResponse {
	choices := make([]ChatCompletionChoice, len(r.Choices))
	for i, choice := range r.Choices {
		choices[i] = ChatCompletionChoice{
			Index: choice.Index,
			Message: ChatCompletionMessage{
				Content: choice.Delta.Content,
				Role:    ChatMessageRoleAssistant,
			},
			FinishReason: choice.FinishReason,
		}
	}
	return ChatCompletionResponse{
		ID:      r.ID,
		Object:  r.Object,
		Created: r.Created,
		Model:   r.Model,
		Choices: choices,
	}
}

type Conversation struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"`
	UserId    string    `json:"user_id"`
	Name      string    `json:"name"`
	Model     string    `json:"model"`
	Summary   string    `json:"summary"`
	ExtraInfo string    `json:"extra_info"`
}

type Message struct {
	Id              string                 `json:"id"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Deleted         bool                   `json:"deleted"`
	UserId          string                 `json:"user_id"`
	ConversationId  string                 `json:"conversation_id"`
	Model           string                 `json:"model"`
	PromptToken     int                    `json:"prompt_token"`
	CompletionToken int                    `json:"completion_token"`
	Description     string                 `json:"description"`
	Request         ChatCompletionRequest  `json:"request"`
	Response        ChatCompletionResponse `json:"response"`
	RawResponse     []byte                 `json:"raw_response"`
}

type Model struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Pricing     struct {
		Prompt     float64 `json:"prompt,omitempty"`
		Completion float64 `json:"completion,omitempty"`
	} `json:"pricing,omitempty"`
	ContextLength   int `json:"context_length,omitempty"`
	PreRequestLimit struct {
		Prompt     int `json:"prompt_tokens,omitempty"`
		Completion int `json:"completion_tokens,omitempty"`
	}
}
