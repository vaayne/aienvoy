package middlerware

import (
	"strings"

	"aienvoy/internal/pkg/config"
	"aienvoy/internal/pkg/dao"
	"aienvoy/internal/pkg/logger"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/daos"
)

// AuthByApiKeyMiddleware is a middleware to auth user by api key
func AuthByApiKeyMiddleware(d *daos.Dao) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			val := c.Get(config.ContextKeyAuthRecord)
			if val == nil {
				auth := c.Request().Header.Get("Authorization")
				if auth != "" {
					auths := strings.Split(auth, " ")
					if len(auths) == 2 && auths[0] == "Bearer" {
						apiKey := auths[1]

						authRecord, err := dao.FindAuthRecordByApiKey(d, apiKey)
						if err != nil {
							logger.SugaredLogger.Infow("error get user by api key", "err", err, "key", auth)
							return apis.NewUnauthorizedError("invalid api key", nil)
						}
						c.Set(config.ContextKeyAuthRecord, authRecord)
						c.Set(config.ContextKeyUserId, authRecord.Id)
						c.Set(config.ContextKeyApiKey, apiKey)
					}
				}
			}
			return next(c)
		}
	}
}
