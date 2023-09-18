package middlerware

import (
	"github.com/Vaayne/aienvoy/internal/pkg/config"

	"github.com/labstack/echo/v5"
	emw "github.com/labstack/echo/v5/middleware"
)

func RequestIDMiddleware() echo.MiddlewareFunc {
	return emw.RequestIDWithConfig(emw.RequestIDConfig{
		TargetHeader: config.ContextKeyRequestId,
		RequestIDHandler: func(c echo.Context, requestID string) {
			c.Set(config.ContextKeyRequestId, requestID)
		},
	})
}
