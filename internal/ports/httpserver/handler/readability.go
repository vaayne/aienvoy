package handler

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo/v5"
	"github.com/vaayne/gtk/cleanweb"
)

var once sync.Once

var parser *cleanweb.Parser

func Readability(c echo.Context) error {
	ctx := c.Request().Context()

	uri := c.QueryParam("uri")
	if uri == "" {
		return c.JSON(http.StatusBadRequest, "uri is required")
	}

	markdown := c.QueryParamDefault("markdown", "false")

	isOutputMarkdown, err := strconv.ParseBool(markdown)
	if err != nil {
		isOutputMarkdown = false
	}

	if parser == nil {
		once.Do(func() {
			parser = cleanweb.NewParser()
		})
	}

	if isOutputMarkdown {
		parser.WithFormatMarkdown()
	}

	article, err := parser.Parse(ctx, uri)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	article.Node = nil
	return c.JSON(http.StatusOK, article)
}
