package claudeweb

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/Vaayne/aienvoy/pkg/session"
	"github.com/google/uuid"
	utls "github.com/refraction-networking/utls"
)

const (
	ModelClaudeWeb   = "claude-2"
	defaultTimezone  = "Asia/Shanghai"
	defaultHost      = "https://claude.ai"
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
)

type Client struct {
	mu         sync.Mutex
	session    *session.Session
	sessionKey string
	orgId      string
	model      string
}

// New will return a Claude request client
func NewClient(sessionKey string) *Client {
	claudeWeb := &Client{
		session:    session.New(session.WithClientHelloID(utls.HelloChrome_100_PSK)),
		sessionKey: sessionKey,
		model:      ModelClaudeWeb,
	}

	orgs, err := claudeWeb.GetOrganizations()
	if err != nil {
		slog.Error("GetOrganizations error", "err", err)
		return nil
	}
	if len(orgs) == 0 {
		slog.Error("GetOrganizations empty")
		return nil
	}
	slog.Info("success get claude org info", "org", orgs[0])
	claudeWeb.orgId = orgs[0].UUID
	return claudeWeb
}

func (cw *Client) GetOrganizations() ([]*Organization, error) {
	uri := fmt.Sprintf("%s/api/organizations", defaultHost)
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	resp, _, err := cw.request(req)
	if err != nil {
		return nil, err
	}

	defer resp.Close()
	var orgs []*Organization
	body, err := io.ReadAll(resp)
	if err != nil {
		return nil, fmt.Errorf("GetOrganizations read response body err: %v", err)
	}
	err = json.Unmarshal(body, &orgs)
	if err != nil {
		slog.Error("Unmarshal error", "body", string(body), "err", err)
		return nil, fmt.Errorf("GetOrganizations unmarshal response body err: %v", err)
	}

	return orgs, nil
}

func (cw *Client) request(req *http.Request) (io.ReadCloser, int, error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.setReqHeaders(req)
	r, err := cw.session.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s %s err: %v", req.Method, req.URL.String(), err)
	}
	if r.StatusCode >= http.StatusBadRequest {
		return nil, r.StatusCode, fmt.Errorf("%s %s err: %v", req.Method, req.URL.String(), r.Status)
	}
	return r.Body, r.StatusCode, nil
}

func (cw *Client) setReqHeaders(req *http.Request) {
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Referer", defaultHost)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Connection", "keep-alive")

	req.AddCookie(&http.Cookie{
		Name:  "sessionKey",
		Value: cw.sessionKey,
	})
}

// GetOrgId will get organization id
func (cw *Client) GetOrgId() string {
	return cw.orgId
}

// GetModel will get default model
func (cw *Client) GetModel() string {
	return cw.model
}

func (cw *Client) ListConversations() ([]*Conversation, error) {
	uri := fmt.Sprintf("%s/api/organizations/%s/chat_conversations", defaultHost, cw.orgId)
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	resp, _, err := cw.request(req)
	if err != nil {
		return nil, err
	}
	var conversations []*Conversation
	body, err := io.ReadAll(resp)
	defer resp.Close()
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
func (cw *Client) GetConversation(id string) (*Conversation, error) {
	uri := fmt.Sprintf("%s/api/organizations/%s/chat_conversations/%s", defaultHost, cw.orgId, id)
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	resp, _, err := cw.request(req)
	if err != nil {
		return nil, err
	}

	var conversation Conversation
	body, err := io.ReadAll(resp)
	defer resp.Close()
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
func (cw *Client) DeleteConversation(id string) error {
	uri := fmt.Sprintf("%s/api/organizations/%s/chat_conversations/%s", defaultHost, cw.orgId, id)
	req, _ := http.NewRequest(http.MethodDelete, uri, nil)
	_, _, err := cw.request(req)
	return err
}

// CreateConversation is used to create conversation
func (cw *Client) CreateConversation(name string) (*Conversation, error) {
	uri := fmt.Sprintf("%s/api/organizations/%s/chat_conversations", defaultHost, cw.orgId)
	params := map[string]any{
		"name": name,
		"uuid": uuid.NewString(),
	}

	paramsBytes, _ := json.Marshal(params)
	req, _ := http.NewRequest(http.MethodPost, uri, bytes.NewReader(paramsBytes))

	resp, statusCode, err := cw.request(req)
	if err != nil {
		return nil, fmt.Errorf("CreateConversation status_code %d err: %v", statusCode, err)
	}

	var conversation Conversation
	body, err := io.ReadAll(resp)
	defer resp.Close()
	if err != nil {
		return nil, fmt.Errorf("CreateConversation read response body err: %v", err)
	}
	err = json.Unmarshal(body, &conversation)
	if err != nil {
		return nil, fmt.Errorf("CreateConversation unmarshal response body err: %v", err)
	}
	slog.Debug("CreateConversation", "status_code", statusCode, "conversation", conversation)
	return &conversation, nil
}

// UpdateConversation is used to update conversation
func (cw *Client) UpdateConversation(id string, name string) error {
	uri := defaultHost + "/api/rename_chat"

	updateReq := UpdateConversationRequest{
		OrganizationUUID: cw.orgId,
		ConversationUUID: id,
		Title:            name,
	}

	paramsBytes, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPost, uri, bytes.NewReader(paramsBytes))
	_, statusCode, err := cw.request(req)
	if err != nil {
		return fmt.Errorf("UpdateConversation status_code %d err: %v", statusCode, err)
	}
	slog.Info("update conversation", "status_code", statusCode)
	return nil
}

func (cw *Client) createChatMessage(id, prompt string) (io.ReadCloser, int, error) {
	uri := defaultHost + "/api/append_message"

	payload := CreateChatMessageRequest{
		Completion: Completion{
			Prompt:   prompt,
			Timezone: defaultTimezone,
			Model:    cw.model,
		},
		OrganizationUUID: cw.orgId,
		ConversationUUID: id,
		Text:             prompt,
		Attachments:      []Attachment{},
	}

	paramsBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, uri, bytes.NewReader(paramsBytes))
	req.Header.Set("Content-Type", "text/event-stream")
	return cw.request(req)
}

func (cw *Client) CreateChatMessage(id, prompt string) (*ChatMessageResponse, error) {
	resp, statusCode, err := cw.createChatMessage(id, prompt)
	if err != nil {
		return nil, fmt.Errorf("CreateChatMessage err: %v", err)
	}

	if statusCode >= http.StatusBadRequest {
		slog.Error("CreateChatMessage", "status_code", statusCode, "text", "")
		return nil, fmt.Errorf("CreateChatMessage status_code %d err: %v", statusCode, err)
	}
	slog.Info("CreateChatMessage", "status_code", statusCode)

	var chatMessageResponse ChatMessageResponse
	sb := strings.Builder{}
	reader := bufio.NewReader(resp)
	defer resp.Close()

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

func (cw *Client) CreateChatMessageStream(id, prompt string, streamChan chan *ChatMessageResponse, errChan chan error) {
	resp, statusCode, err := cw.createChatMessage(id, prompt)
	if err != nil {
		errChan <- fmt.Errorf("CreateChatMessage failed with status_code %d, err: %v", statusCode, err)
		return
	}

	if statusCode >= http.StatusBadRequest {
		slog.Error("CreateChatMessageStream", "status_code", statusCode, "text", "")
		errChan <- fmt.Errorf("CreateChatMessage status_code %d err: %v", statusCode, err)
		return
	}
	slog.Info("CreateChatMessageStream", "status_code", statusCode)

	reader := bufio.NewReader(resp)
	defer resp.Close()

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
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
			streamChan <- &chatMessageResponse
		}
	}
}
