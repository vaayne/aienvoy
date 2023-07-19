package middlerware

import (
	"aienvoy/internal/pkg/config"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/daos"
)

func DaoMiddleware(d *daos.Dao) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(config.ContextKeyDao, d)
			return next(c)
		}
	}
}
