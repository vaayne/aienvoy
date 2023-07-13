package httpserver

import (
	"net/http"

	"openai-dashboard/internal/ports/httpserver/handler"
	"openai-dashboard/internal/ports/httpserver/middlerware"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
)

func RegisterRoutes(e *echo.Echo, app *pocketbase.PocketBase) {
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/login.html")
	})
	e.Static("/*", "web")

	v1 := e.Group("/v1", middlerware.AuthByApiKeyMiddleware(app.Dao()), apis.RequireAdminOrRecordAuth())
	openaiHandler := handler.NewOpenAIHandler()
	v1.POST("/chat/completions", openaiHandler.Chat)
	v1.POST("/embeddings", openaiHandler.CreateEmbeddings)
	v1.GET("/status", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
}
