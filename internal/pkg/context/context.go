package context

import (
	"context"

	"aienvoy/internal/pkg/config"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type Context struct {
	context.Context
}

func New(ctx context.Context) Context {
	return Context{ctx}
}

func FromEchoContext(c echo.Context) Context {
	return Context{c.Request().Context()}
}

func (c *Context) User() *models.Record {
	if user, ok := c.Value(config.ContextKeyAuthRecord).(*models.Record); ok {
		return user
	}
	return nil
}

func (c *Context) APIKey() string {
	if apiKey, ok := c.Value(config.ContextKeyApiKey).(string); ok {
		return apiKey
	}
	return ""
}

func (c *Context) RequestId() string {
	if requestId, ok := c.Value(config.ContextKeyRequestId).(string); ok {
		return requestId
	}
	return ""
}

func (c *Context) UserId() string {
	if userId, ok := c.Value(config.ContextKeyUserId).(string); ok {
		return userId
	}
	return ""
}

func (c *Context) Dao() *daos.Dao {
	if dao, ok := c.Value(config.ContextKeyDao).(*daos.Dao); ok {
		return dao
	}
	return nil
}
