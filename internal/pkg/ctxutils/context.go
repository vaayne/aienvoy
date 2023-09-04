package ctxutils

import (
	"context"

	"aienvoy/internal/pkg/config"

	"github.com/pocketbase/pocketbase/daos"
)

func GetDao(ctx context.Context) *daos.Dao {
	val, ok := ctx.Value(config.ContextKeyDao).(*daos.Dao)
	if !ok {
		return nil
	}
	return val
}

func GetUserId(ctx context.Context) string {
	return getString(ctx, config.ContextKeyUserId)
}

func GetApiKey(ctx context.Context) string {
	return getString(ctx, config.ContextKeyApiKey)
}

func GetRequestId(ctx context.Context) string {
	return getString(ctx, config.ContextKeyRequestId)
}

func getString(ctx context.Context, key string) string {
	val, ok := ctx.Value(key).(string)
	if !ok {
		return ""
	}
	return val
}
