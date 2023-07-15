package middlerware

import (
	"strings"

	"aienvoy/internal/pkg/logger"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/daos"
)

const ApiKeyCollectionName = "api_keys"

// AuthByApiKeyMiddleware is a middleware to auth user by api key
func AuthByApiKeyMiddleware(dao *daos.Dao) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			val := c.Get(apis.ContextAuthRecordKey)
			if val == nil {
				auth := c.Request().Header.Get("Authorization")
				if auth != "" {
					auths := strings.Split(auth, " ")
					if len(auths) == 2 && auths[0] == "Bearer" {
						apiKey := auths[1]
						authRecord, err := dao.FindFirstRecordByData(ApiKeyCollectionName, "key", apiKey)
						if err != nil {
							logger.SugaredLogger.Infow("error get user by api key", "err", err, "key", auth)
							return apis.NewUnauthorizedError("invalid api key", nil)
						}
						c.Set(apis.ContextAuthRecordKey, authRecord)
					}
				}
			}
			return next(c)
		}
	}
}
