package httpserver

import (
	"embed"
	"net/http"

	"github.com/Vaayne/aienvoy/internal/ports/httpserver/handler"
	"github.com/Vaayne/aienvoy/internal/ports/httpserver/middlerware"

	"github.com/labstack/echo/v5"
	emw "github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
)

func RegisterRoutes(e *echo.Echo, app *pocketbase.PocketBase, staticFiles embed.FS) {
	// web static files
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/web/")
	})
	e.Add(http.MethodGet, "/web/*", echo.WrapHandler(http.FileServer(http.FS(staticFiles))))

	// v1 apis
	mds := []echo.MiddlewareFunc{
		middlerware.ContextMiddleware(),
		middlerware.RequestIDMiddleware(),
		middlerware.AuthByApiKeyMiddleware(app.Dao()),
		apis.RequireAdminOrRecordAuth(),
		middlerware.DaoMiddleware(app.Dao()),
		middlerware.LoggerMiddleware(),
		emw.CORS(),
	}

	v1 := e.Group("/v1", mds...)
	openaiHandler := handler.NewOpenAIHandler()
	v1.POST("/chat/completions", openaiHandler.Chat)
	v1.POST("/embeddings", openaiHandler.CreateEmbeddings)
	v1.GET("/models", openaiHandler.GetModels)
	v1.GET("/status", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
}
