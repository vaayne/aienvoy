package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/bincooo/requests"
	"github.com/bincooo/requests/models"
	rurl "github.com/bincooo/requests/url"
	"github.com/tidwall/gjson"
)

const (
	BASE_URI = "https://claude.ai"
	JA3      = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0"
	UA       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0"
)

// MixMap is a type alias for map[string]interface{}
type MixMap = map[string]interface{}

// Client is a ChatGPT request client
type Client struct {
	opts    Options // custom options
	httpCli *http.Client
}

// NewClient will return a ChatGPT request client
func NewClient(options ...Option) *Client {
	cli := &Client{
		opts: Options{
			BaseUri:   BASE_URI,
			UserAgent: UA,
			Timeout:   120 * time.Second, // set default timeout
			Model:     "claude-2",        // set default chat model
			Debug:     false,
		},
	}

	// load custom options
	for _, option := range options {
		option(cli)
	}

	cli.initHttpClient()

	return cli
}

func (c *Client) initHttpClient() {
	transport := &http.Transport{}
	if c.opts.Proxy != "" {
		proxy, err := url.Parse(c.opts.Proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}

	c.httpCli = &http.Client{
		Timeout:   c.opts.Timeout,
		Transport: transport,
	}
}

// Get will request api with Get method
func (c *Client) Get(uri string) (*gjson.Result, error) {
	if !strings.HasPrefix(uri, "http") {
		uri = c.opts.BaseUri + uri
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("new request failed: %v", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return c.parseBody(resp.Body)
}

// Post will request api with Post method
func (c *Client) Post(uri string, params MixMap) (*models.Response, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %v", err)
	}

	if !strings.HasPrefix(uri, "http") {
		uri = c.opts.BaseUri + uri
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("new request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req)
}

// Delete will request api with delete method
func (c *Client) Delete(uri string, params MixMap) (*gjson.Result, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %v", err)
	}

	if !strings.HasPrefix(uri, "http") {
		uri = c.opts.BaseUri + uri
	}

	req, err := http.NewRequest(http.MethodDelete, uri, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("new request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	return c.parseBody(resp.Body)
}

func (c *Client) doRequest(req *http.Request) (resp *models.Response, err error) {
	headers := rurl.NewHeaders()
	headers.Set("User-Agent", UA)
	headers.Set("Accept-Language", "en-US,en;q=0.5")
	headers.Set("Referer", "https://claude.ai/chats")
	headers.Set("Content-Type", "application/json")
	headers.Set("Sec-Fetch-Dest", "empty")
	headers.Set("Sec-Fetch-Mode", "cors")
	headers.Set("Sec-Fetch-Site", "same-origin")
	headers.Set("Connection", "keep-alive")
	headers.Set("Cookie", fmt.Sprintf("sessionKey=%s", c.opts.SessionKey))

	if c.opts.Debug {
		reqInfo, _ := httputil.DumpRequest(req, true)
		log.Printf("http request info: \n%s\n", reqInfo)
	}
	newReq := rurl.NewRequest()
	newReq.Headers = headers
	newReq.Timeout = c.opts.Timeout
	newReq.Ja3 = JA3

	rawUrl := req.URL.String()
	// logger.SugaredLogger.Debugw("request", "url", rawUrl, "headers", newReq.Headers, "ja3", newReq.Ja3)

	switch req.Method {
	case http.MethodGet:
		resp, err = requests.Get(rawUrl, newReq)
	default:
		resp, err = requests.RequestStream(http.MethodPost, rawUrl, newReq)
	}
	return resp, err
}

func (c *Client) parseBody(resp io.ReadCloser) (*gjson.Result, error) {
	body, err := io.ReadAll(resp)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	res := gjson.ParseBytes(body)

	return &res, nil
}

// GetOrgid will get orgid set in option
func (c *Client) GetOrgid() string {
	return c.opts.Orgid
}

// GetModel will get model set in option
func (c *Client) GetModel() string {
	return c.opts.Model
}
