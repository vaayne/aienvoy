package claudeweb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
)

const (
	BASE_URI = "https://claude.ai"
	UA       = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
	JA3      = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21,29-23-24,0"
)

// MixMap is a type alias for map[string]interface{}
type MixMap = map[string]interface{}

func NewMixMap(in interface{}) MixMap {
	m := make(MixMap)
	inrec, _ := json.Marshal(in)
	json.Unmarshal(inrec, &m)
	return m
}

// Client is a ChatGPT request client
type Client struct {
	opts    Options // custom options
	req     *url.Request
	httpCli *http.Client
}

// NewClient will return a ChatGPT request client
func NewClient(options ...Option) *Client {
	cli := &Client{
		opts: Options{
			BaseUri:   BASE_URI,
			UserAgent: UA,
			JA3:       JA3,
			Timeout:   120 * time.Second, // set default timeout
			Model:     "claude-2",        // set default chat model
			Debug:     false,
		},
	}

	// load custom options
	for _, option := range options {
		option(cli)
	}

	cli.req = cli.newReq()

	return cli
}

func (c *Client) newReq() *url.Request {
	req := url.NewRequest()

	// set headers
	headers := url.NewHeaders()
	headers.Set("User-Agent", UA)
	headers.Set("Accept-Language", "en-US,en;q=0.5")
	headers.Set("Referer", c.opts.BaseUri)
	headers.Set("Content-Type", "application/json")
	headers.Set("Sec-Fetch-Dest", "empty")
	headers.Set("Sec-Fetch-Mode", "cors")
	headers.Set("Sec-Fetch-Site", "same-origin")
	headers.Set("Connection", "keep-alive")
	req.Headers = headers

	// set cookies
	cookies := map[string]interface{}{
		"sessionKey": c.opts.SessionKey,
	}
	req.Cookies = url.ParseCookies(c.opts.BaseUri, cookies)

	// set ja3
	req.Ja3 = c.opts.JA3

	if c.opts.Proxy != "" {
		req.Proxies = c.opts.Proxy
	}

	req.Timeout = c.opts.Timeout
	return req
}

func (c *Client) Get(uri string) (*models.Response, error) {
	if !strings.HasPrefix(uri, "http") {
		uri = c.opts.BaseUri + uri
	}
	r, err := requests.Get(uri, c.req)
	if err != nil {
		return nil, fmt.Errorf("GET %s err: %v", uri, err)
	}
	return r, nil
}

func (c *Client) Post(uri string, params MixMap, headers map[string]string) (*models.Response, error) {
	if !strings.HasPrefix(uri, "http") {
		uri = c.opts.BaseUri + uri
	}
	req := c.newReq()
	req.Json = params
	if headers != nil {
		for k, v := range headers {
			req.Headers.Set(k, v)
		}
	}

	// logger.SugaredLogger.Debugw("request", "request", req)

	r, err := requests.Post(uri, req)
	if err != nil {
		return nil, fmt.Errorf("POST %s err: %v", uri, err)
	}
	return r, nil
}

// Delete will request api with delete method
func (c *Client) Delete(uri string) (*models.Response, error) {
	if !strings.HasPrefix(uri, "http") {
		uri = c.opts.BaseUri + uri
	}
	r, err := requests.Delete(uri, c.req)
	if err != nil {
		return nil, fmt.Errorf("DELETE %s err: %v", uri, err)
	}
	return r, nil
}

// GetOrgid will get orgid set in option
func (c *Client) GetOrgid() string {
	return c.opts.Orgid
}

// GetModel will get model set in option
func (c *Client) GetModel() string {
	return c.opts.Model
}

// GetOrganizations is used to get account organizations
func (c *Client) GetOrganizations() (*models.Response, error) {
	uri := "/api/organizations"

	return c.Get(uri)
}

// Options for request client
type Options struct {
	// Debug is used to output debug message
	Debug bool
	// Timeout is used to end http request after timeout duration
	Timeout time.Duration
	// Proxy is used to proxy request
	Proxy string
	// SessionKey is used to set authorization key
	SessionKey string
	// Model is the chat model
	Model string
	// BaseUri is the api base uri
	BaseUri string
	// UserAgent is used to set user agent in header
	UserAgent string
	// Orgid is user's uuid
	Orgid string
	// JA3 is used to set ja3
	JA3 string
}

// Option is used to set custom option
type Option func(*Client)

// WithDebug is used to output debug message
func WithDebug(debug bool) Option {
	return func(c *Client) {
		c.opts.Debug = debug
	}
}

// WithTimeout is used to set request timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.opts.Timeout = timeout
	}
}

// WithProxy is used to set request proxy
func WithProxy(proxy string) Option {
	return func(c *Client) {
		c.opts.Proxy = proxy
	}
}

// WithSessionKey is used to set session key in cookie
func WithSessionKey(sessionKey string) Option {
	return func(c *Client) {
		c.opts.SessionKey = sessionKey
	}
}

// WithModel is used to set chat model
func WithModel(model string) Option {
	return func(c *Client) {
		c.opts.Model = model
	}
}

// WithBaseUri is used to set api base uri
func WithBaseUri(baseUri string) Option {
	return func(c *Client) {
		c.opts.BaseUri = baseUri
	}
}

// WithUserAgent is used to set user_agent
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.opts.UserAgent = userAgent
	}
}

// WithOrgid is used to set orgid
func WithOrgid(orgid string) Option {
	return func(c *Client) {
		c.opts.Orgid = orgid
	}
}

func WithJA3(ja3 string) Option {
	return func(c *Client) {
		c.opts.JA3 = ja3
	}
}
