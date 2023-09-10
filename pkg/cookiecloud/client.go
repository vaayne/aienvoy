package cookiecloud

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
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
	client       *http.Client
	key          string
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
		client:       http.DefaultClient,
		cacheService: cache.New(10*time.Minute, time.Hour),
	}
	c.key = c.getKey()
	return c
}

func (c *CookieCloud) fetchData() (*CookieData, error) {
	resp, err := c.client.Get(c.Host + "/get/" + c.UUID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var encryptedData struct {
		Encrypted string `json:"encrypted"`
	}
	// var encryptedData map[string]any
	err = json.NewDecoder(resp.Body).Decode(&encryptedData)
	if err != nil {
		slog.Error("decode err", err)
		return nil, err
	}

	data, err := decryptData(c.key, encryptedData.Encrypted)
	if err != nil {
		return nil, err
	}
	return data, nil
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

func (c *CookieCloud) getKey() string {
	hash := md5.Sum([]byte(c.UUID + "-" + c.Password))
	key := hex.EncodeToString(hash[:])[:16]
	slog.Debug("decryption key", "key", key)
	return key
}

func decryptData(passphrase, encryptedData string) (*CookieData, error) {
	// base64 decode encrypted data
	encrypted, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("base64 decode err: %w", err)
	}

	salt := encrypted[8:16]
	key_iv, _ := bytesToKey([]byte(passphrase), salt, 48)
	key := key_iv[:32]
	iv := key_iv[32:]
	// fmt.Printf("key: %x\niv: %x\nkey: %x\nkey_iv: %x\n", key, iv, key, key_iv)

	ciphertext := encrypted[16:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes new cipher err: %w", err)
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)
	// slog.Info("decrypt success", "ciphertext", string(ciphertext))

	var decryptedData CookieData

	ciphertext = bytes.TrimRightFunc(ciphertext, func(r rune) bool {
		return r != '}'
	})

	err = json.Unmarshal(ciphertext, &decryptedData)
	if err != nil {
		slog.Error("unmarshal error", "data", ciphertext)
		return nil, fmt.Errorf("unmarshal err: %w", err)
	}
	// fmt.Printf("%v", decryptedData)
	return &decryptedData, nil
}

func bytesToKey(data []byte, salt []byte, output int) ([]byte, error) {
	if len(salt) != 8 {
		return nil, fmt.Errorf("expected salt of length 8, got %d", len(salt))
	}
	data = append(data, salt...)
	hash := md5.Sum(data)
	key := hash[:]
	finalKey := append([]byte(nil), key...)
	for len(finalKey) < output {
		hash = md5.Sum(append(key, data...))
		key = hash[:]
		finalKey = append(finalKey, key...)
	}
	return finalKey[:output], nil
}
