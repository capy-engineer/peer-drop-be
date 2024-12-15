package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	httpservice "peer-drop/internal/adapters/http"
)

func InitServer() *echo.Echo {
	e := echo.New()

	// Attach middlewares.
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Define routes.
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	e.GET("/ws", httpservice.SignalingHandler)
	e.GET("/connect", httpservice.ConnectHandler)

	return e
}
