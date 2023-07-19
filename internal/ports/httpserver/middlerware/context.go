package middlerware

import (
	"context"

	"github.com/labstack/echo/v5"
)

func ContextMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(contextValue{c})
		}
	}
}

type contextValue struct {
	echo.Context
}

// Get retrieves data from the context.
func (ctx contextValue) Get(key string) interface{} {
	// get old context value
	val := ctx.Context.Get(key)
	if val != nil {
		return val
	}
	return ctx.Request().Context().Value(key)
}

// Set saves data in the context.
func (ctx contextValue) Set(key string, val interface{}) {
	ctx.Context.Set(key, val)
	// nolint:staticcheck
	ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), key, val)))
}
