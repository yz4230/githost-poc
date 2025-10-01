package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

func RegisterMisc(injector *do.Injector, e *echo.Echo) {
	e.GET("/api/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
}
