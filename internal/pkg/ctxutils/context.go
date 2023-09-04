package ctxutils

import (
	"context"

	"aienvoy/internal/pkg/config"

	"github.com/pocketbase/pocketbase/daos"
)

func GetDao(ctx context.Context) *daos.Dao {
	return ctx.Value(config.ContextKeyDao).(*daos.Dao)
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
	return ctx.Value(key).(string)
}
