package claude

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/tidwall/gjson"
)

// ChatStream: chat strem reply
type ChatStream struct {
	Stream chan *gjson.Result // chat message stream
	Err    error              // error message
}

// GetChatStream is used to get chat stream
func (c *Client) GetChatStream(params MixMap) (*ChatStream, error) {
	uri := "/api/append_message"

	resp, err := c.Post(uri, params)
	if err != nil {
		return nil, err
	}

	contentType := resp.Headers.Get("Content-Type")
	// not event-strem response
	if !strings.HasPrefix(contentType, "text/event-stream") {
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if strings.HasPrefix(contentType, "application/json") {
			res := gjson.ParseBytes(body)
			errmsg := res.Get("error.message").String()
			errcode := res.Get("error.type").String()
			if errmsg != "" {
				return nil, fmt.Errorf("response failed: [%s] %s", errcode, errmsg)
			}
		}

		return nil, fmt.Errorf("response failed: [%d] %s", resp.StatusCode, body)
	}

	chatStream := &ChatStream{
		Stream: make(chan *gjson.Result),
		Err:    nil,
	}

	go func() {
		defer resp.Body.Close()
		defer close(chatStream.Stream)
		scanner := bufio.NewScanner(resp.Body)
		for {
			line := scanner.Bytes()

			if len(line) < 6 {
				continue
			}

			if bytes.HasPrefix(line, []byte("data: [DONE]")) {
				return
			}

			jres := gjson.ParseBytes([]byte(line[6:]))

			if jres.Get("model").String() == "" {
				chatStream.Err = fmt.Errorf("invalid stream data: %s", line)
				return
			}

			chatStream.Stream <- &jres
		}
	}()

	return chatStream, nil
}

// GetConversations is used to get conversations
func (c *Client) ListConversations() (*gjson.Result, error) {
	uri := fmt.Sprintf("/api/organizations/%s/chat_conversations", c.GetOrgid())

	return c.Get(uri)
}

// NewConversation is used to new conversation
func (c *Client) NewConversation(params MixMap) (*gjson.Result, error) {
	uri := fmt.Sprintf("/api/organizations/%s/chat_conversations", c.opts.Orgid)

	resp, err := c.Post(uri, params)
	if err != nil {
		return nil, err
	}
	return c.parseBody(resp.Body)
}

// DelConversation is used to del conversation
func (c *Client) DelConversation(conversationUuid string) (*gjson.Result, error) {
	uri := fmt.Sprintf("/api/organizations/%s/chat_conversations/%s", c.opts.Orgid, conversationUuid)

	return c.Delete(uri, nil)
}

// GetOrganizations is used to get account organizations
func (c *Client) GetOrganizations() (*gjson.Result, error) {
	uri := "/api/organizations"

	return c.Get(uri)
}
