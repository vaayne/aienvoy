package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var DefaultClient = cache.New(1*time.Hour, 24*time.Hour)

// CacheFunc cache func result
// func input could be any params
// func output should be (any, error)
func CacheFunc(fn func(params ...any) (any, error), cacheKey string, duration time.Duration, params ...any) (any, error) {
	val, ok := DefaultClient.Get(cacheKey)
	if ok {
		return val, nil
	}
	val, err := fn(params...)
	if err != nil {
		return nil, err
	}
	DefaultClient.Set(cacheKey, val, duration)
	return val, err
}
