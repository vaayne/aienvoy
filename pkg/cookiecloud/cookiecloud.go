package cookiecloud

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	cookieCloudCacheKey = "cookieCloudCacheKey"
	cacheDuration       = 10 * time.Minute
)

type CacheService interface {
	Set(k string, x interface{}, d time.Duration)
	Get(k string) (interface{}, bool)
}

type CookieCloud struct {
	Host         string
	UUID         string
	Password     string
	cacheService CacheService
}

type Cookie struct {
	Name           string
	Value          string
	Domain         string
	Path           string
	ExpirationDate float64
	HostOnly       bool
	HttpOnly       bool
	Session        bool
	Secure         bool
	StoreId        string
	SameSite       string
}

type CookieData struct {
	CookieData       map[string][]Cookie `json:"cookie_data"`
	LocalStorateData map[string]any      `json:"local_storage_data"`
}

func New(host, uuid, pass string) *CookieCloud {
	c := &CookieCloud{
		Host:         host,
		UUID:         uuid,
		Password:     pass,
		cacheService: cache.New(10*time.Minute, time.Hour),
	}
	return c
}

func (c *CookieCloud) GetCookies(domain string) ([]Cookie, error) {
	var cookieData *CookieData
	var err error
	val, ok := c.cacheService.Get(cookieCloudCacheKey)
	if ok {
		cookieData = val.(*CookieData)
	} else {
		cookieData, err = c.fetchData()
		if err != nil {
			return nil, err
		}
		c.cacheService.Set(cookieCloudCacheKey, cookieData, cacheDuration)
	}

	return cookieData.CookieData[domain], nil
}

// GetHttpCookies get http cookies for domain
func (c *CookieCloud) GetHttpCookies(domain string) ([]*http.Cookie, error) {
	cookies := make([]*http.Cookie, 0)
	cks, err := c.GetCookies(domain)
	if err != nil {
		return cookies, err
	}
	for _, ck := range cks {
		cookies = append(cookies, &http.Cookie{
			Name:  ck.Name,
			Value: ck.Value,
		})
	}
	return cookies, nil
}

func (c *CookieCloud) GetCookie(domain, key string) (Cookie, error) {
	ck := Cookie{}
	cookies, err := c.GetCookies(domain)
	if err != nil {
		return ck, err
	}
	for _, cookie := range cookies {
		if cookie.Name == key {
			return cookie, err
		}
	}
	return ck, nil
}

func (c *CookieCloud) fetchData() (*CookieData, error) {
	resp, err := http.DefaultClient.Get(c.Host + "/get/" + c.UUID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var encryptedData struct {
		Encrypted string `json:"encrypted"`
	}
	err = json.NewDecoder(resp.Body).Decode(&encryptedData)
	if err != nil {
		slog.Error("decode err", "err", err)
		return nil, err
	}

	data, err := Decrypt(c.getKey(), encryptedData.Encrypted)
	if err != nil {
		return nil, err
	}
	return parseResponse(data)
}

func (c *CookieCloud) getKey() string {
	hash := md5.Sum([]byte(c.UUID + "-" + c.Password))
	key := hex.EncodeToString(hash[:])[:16]
	return key
}

func parseResponse(data []byte) (*CookieData, error) {
	var decryptedData CookieData
	err := json.Unmarshal(data, &decryptedData)
	if err != nil {
		slog.Error("unmarshal error", "data", data)
		return nil, fmt.Errorf("unmarshal err: %w", err)
	}
	return &decryptedData, nil
}
