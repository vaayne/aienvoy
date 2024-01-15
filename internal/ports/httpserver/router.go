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
	mds := []echo.MiddlewareFunc{
		middlerware.ContextMiddleware(),
		middlerware.RequestIDMiddleware(),
		middlerware.DaoMiddleware(app.Dao()),
		middlerware.LoggerMiddleware(),
		emw.CORS(),
	}

	e.Use(mds...)

	// web static files
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/web/")
	})
	e.Add(http.MethodGet, "/web/*", echo.WrapHandler(http.FileServer(http.FS(staticFiles))))

	// v1 apis
	v1 := e.Group("/v1", middlerware.AuthByApiKeyMiddleware(app.Dao()), apis.RequireAdminOrRecordAuth())
	llmHandler := handler.NewLLMHandler()
	v1.POST("/chat/completions", llmHandler.CreateChatCompletion)
	// v1.POST("/embeddings", llmHandler.CreateEmbeddings)
	// v1.GET("/models", llmHandler.GetModels)
	v1.GET("/status", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// conversation
	v1.POST("/conversations", llmHandler.CreateConversation)
	v1.GET("/conversations", llmHandler.ListConversations)
	v1.GET("/conversations/:id", llmHandler.GetConversation)
	v1.DELETE("/conversations/:id", llmHandler.DeleteConversation)

	// converation message
	v1.POST("/conversations/:id/messages", llmHandler.CreateMessage)
	v1.GET("/conversations/:conversationId/messages", llmHandler.ListMessages)
	v1.GET("/conversations/:conversationId/messages/:messageId", llmHandler.GetMessage)
	v1.DELETE("/conversations/:conversationId/messages/:messageId", llmHandler.DeleteMessage)

	// read article using readability
	e.GET("/readability", handler.Readability)
}
