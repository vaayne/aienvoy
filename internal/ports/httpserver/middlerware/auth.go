package middlerware

import (
	"net/http"

	"openai-dashboard/internal/pkg/logger"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/daos"
)

const (
	ApiKeyHeader         string = "X-API-KEY"
	ApiKeyCollectionName string = "api_keys"
)

// AuthByApiKeyMiddleware
func AuthByApiKeyMiddleware(dao *daos.Dao) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get(ApiKeyHeader)
			if auth != "" {
				record, err := dao.FindFirstRecordByData(ApiKeyCollectionName, "key", auth)
				if err != nil {
					logger.SugaredLogger.Infow("error get user by api key", "err", err, "key", auth)
					return apis.NewUnauthorizedError("invalid api key", nil)
				}
				userId := record.GetString("user_id")
				authRecord, err := dao.FindRecordById("users", userId)
				if err != nil {
					logger.SugaredLogger.Errorw("error get user by id", "err", err, "user_id", userId)
					return apis.NewApiError(http.StatusInternalServerError, "api key auth error", nil)
				}
				c.Set(apis.ContextAuthRecordKey, authRecord)
			}
			return next(c)
		}
	}
}
