package claudeweb

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/wangluozhe/requests/models"
)

const (
	DEFAULT_MODEL    = "claude-2"
	DEFAULT_TIMEZONE = "Asia/Shanghai"
)

type ClaudeWeb struct {
	Client
	orgId string
}

// NewClaudeWeb returns a new ClaudeWeb client
func NewClaudeWeb(token string, opts ...Option) *ClaudeWeb {
	opts = append(opts, WithSessionKey(token))
	client := &ClaudeWeb{
		Client: *NewClient(opts...),
	}
	orgs, err := client.GetOrganizations()
	if err != nil {
		slog.Error("get organization err", "err", err)
		return nil
	}
	if len(orgs) == 0 {
		slog.Error("no organization found")
		return nil
	}
	client.orgId = orgs[0].UUID
	slog.Info(fmt.Sprintf("org info: %v", orgs[0]))
	return client
}

func (c *ClaudeWeb) GetOrganizations() ([]*Organization, error) {
	uri := "/api/organizations"

	resp, err := c.Get(uri)
	if err != nil {
		return nil, err
	}

	var orgs []*Organization
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("GetOrganizations read response body err: %v", err)
	}
	err = json.Unmarshal(body, &orgs)
	if err != nil {
		return nil, fmt.Errorf("GetOrganizations unmarshal response body err: %v", err)
	}

	return orgs, nil
}

func (c *ClaudeWeb) ListConversations() ([]*Conversation, error) {
	uri := fmt.Sprintf("/api/organizations/%s/chat_conversations", c.orgId)

	resp, err := c.Get(uri)
	if err != nil {
		return nil, err
	}

	var conversations []*Conversation
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("ListConversations read response body err: %v", err)
	}
	err = json.Unmarshal(body, &conversations)
	if err != nil {
		return nil, fmt.Errorf("ListConversations unmarshal response body err: %v", err)
	}

	slog.Debug("conversations", "conversations", conversations)

	return conversations, nil
}

// GetConversation is used to get conversation
func (c *ClaudeWeb) GetConversation(id string) (*Conversation, error) {
	uri := fmt.Sprintf("/api/organizations/%s/chat_conversations/%s", c.orgId, id)
	resp, err := c.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("GetConversation err: %v", err)
	}

	var conversation Conversation
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("GetConversation read response body err: %v", err)
	}
	err = json.Unmarshal(body, &conversation)
	if err != nil {
		return nil, fmt.Errorf("GetConversation unmarshal response body err: %v", err)
	}

	return &conversation, nil
}

// DeleteConversation is used to delete conversation
func (c *ClaudeWeb) DeleteConversation(id string) error {
	uri := fmt.Sprintf("/api/organizations/%s/chat_conversations/%s", c.orgId, id)
	_, err := c.Delete(uri)
	if err != nil {
		return fmt.Errorf("DeleteConversation err: %v", err)
	}

	return nil
}

// CreateConversation is used to create conversation
func (c *ClaudeWeb) CreateConversation(name string) (*Conversation, error) {
	uri := fmt.Sprintf("/api/organizations/%s/chat_conversations", c.orgId)
	params := MixMap{
		"name": name,
		"uuid": uuid.NewString(),
	}
	resp, err := c.Post(uri, params, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateConversation err: %v", err)
	}

	var conversation Conversation
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("CreateConversation read response body err: %v", err)
	}
	err = json.Unmarshal(body, &conversation)
	if err != nil {
		return nil, fmt.Errorf("CreateConversation unmarshal response body err: %v", err)
	}
	slog.Debug("CreateConversation", "status_code", resp.StatusCode, "conversation", conversation)
	return &conversation, nil
}

// UpdateConversation is used to update conversation
func (c *ClaudeWeb) UpdateConversation(id string, name string) error {
	uri := "/api/rename_chat"

	updateReq := UpdateConversationRequest{
		OrganizationUUID: c.orgId,
		ConversationUUID: id,
		Title:            name,
	}

	params := NewMixMap(updateReq)
	resp, err := c.Post(uri, params, nil)
	if err != nil {
		return fmt.Errorf("UpdateConversation status_code %d err: %v", resp.StatusCode, err)
	}
	slog.Info("update conversation", "status_code", resp.StatusCode)
	return nil
}

func (c *ClaudeWeb) createChatMessage(id, prompt string) (*models.Response, error) {
	uri := "/api/append_message"

	payload := CreateChatMessageRequest{
		Completion: Completion{
			Prompt:   prompt,
			Timezone: DEFAULT_TIMEZONE,
			Model:    DEFAULT_MODEL,
		},
		OrganizationUUID: c.orgId,
		ConversationUUID: id,
		Text:             prompt,
		Attachments:      []Attachment{},
	}

	params := NewMixMap(payload)
	headers := map[string]string{
		"Content-Type": "text/event-stream",
	}
	return c.Post(uri, params, headers)
}

func (c *ClaudeWeb) CreateChatMessage(id, prompt string) (*ChatMessageResponse, error) {
	resp, err := c.createChatMessage(id, prompt)
	if err != nil {
		return nil, fmt.Errorf("CreateChatMessage err: %v", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		slog.Error("CreateChatMessage", "status_code", resp.StatusCode, "text", resp.Text)
		return nil, fmt.Errorf("CreateChatMessage status_code %d err: %v", resp.StatusCode, err)
	}
	slog.Info("CreateChatMessage", "status_code", resp.StatusCode)

	var chatMessageResponse ChatMessageResponse
	sb := strings.Builder{}
	reader := bufio.NewReader(resp.Body)

	defer resp.Body.Close()

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("CreateChatMessage read response body err: %v", err)
		}
		if len(line) > 6 {
			err = json.Unmarshal(line[6:], &chatMessageResponse)
			if err != nil {
				return nil, fmt.Errorf("CreateChatMessage unmarshal response body err: %v", err)
			}
			sb.WriteString(chatMessageResponse.Completion)
		}
	}
	chatMessageResponse.Completion = sb.String()
	return &chatMessageResponse, nil
}

func (c *ClaudeWeb) CreateChatMessageStream(id, prompt string, streamChan chan *ChatMessageResponse, errChan chan error) {
	c.CreateChatMessageStreamWithFullResponse(id, prompt, streamChan, nil, errChan)
}

func (c *ClaudeWeb) CreateChatMessageStreamWithFullResponse(id, prompt string, streamChan chan *ChatMessageResponse, fullRespChan chan string, errChan chan error) {
	resp, err := c.createChatMessage(id, prompt)
	if err != nil {
		errChan <- fmt.Errorf("CreateChatMessage err: %v", err)
		return
	}

	if resp.StatusCode >= http.StatusBadRequest {
		slog.Error("CreateChatMessageStream", "status_code", resp.StatusCode, "text", resp.Text)
		errChan <- fmt.Errorf("CreateChatMessage status_code %d err: %v", resp.StatusCode, err)
		return
	}
	slog.Info("CreateChatMessageStream", "status_code", resp.StatusCode)

	reader := bufio.NewReader(resp.Body)

	defer resp.Body.Close()
	fullResp := strings.Builder{}

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				if fullRespChan != nil {
					fullRespChan <- fullResp.String()
				}
				slog.Info("done with CreateChatMessageStream", "cov_id", id)
				errChan <- io.EOF
				return
			}
			errChan <- fmt.Errorf("createChatMessageStream read response body err: %v", err)
			return
		}

		if len(line) > 6 {
			var chatMessageResponse ChatMessageResponse
			err = json.Unmarshal(line[6:], &chatMessageResponse)
			if err != nil {
				errChan <- fmt.Errorf("createChatMessageStream unmarshal response body err: %v", err)
				return
			}
			if fullRespChan != nil {
				fullResp.WriteString(chatMessageResponse.Completion)
			}
			streamChan <- &chatMessageResponse
		}
	}
}
