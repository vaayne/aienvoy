package middlerware

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	emw "github.com/labstack/echo/v5/middleware"
)

func LoggerMiddleware() echo.MiddlewareFunc {
	return emw.RequestLoggerWithConfig(emw.RequestLoggerConfig{
		LogStatus:        true,
		LogError:         true,
		LogLatency:       true,
		LogMethod:        true,
		LogURI:           true,
		LogURIPath:       true,
		LogUserAgent:     true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogValuesFunc: func(c echo.Context, v emw.RequestLoggerValues) error {
			slog.InfoContext(c.Request().Context(),
				"request finished",
				"start_time", v.StartTime,
				"duration_ms", v.Latency.Milliseconds(),
				"duration", v.Latency.Seconds(),
				"method", v.Method,
				"url", v.URI,
				"path", v.RoutePath,
				"user_agent", v.UserAgent,
				"error", v.Error,
				"content_length", v.ContentLength,
				"response_size", v.ResponseSize,
				"status", v.Status,
			)
			return nil
		},
	})
}
