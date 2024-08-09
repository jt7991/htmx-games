package main

import (
	"madlibs-htmx/handlers"
	"madlibs-htmx/ws"

	"github.com/labstack/echo/v4"
)

func main() {
	hub := ws.NewHub()
	go hub.Run()

	e := echo.New()

	setupRoutes(e, hub)
	e.Logger.Fatal(e.Start(":1323"))
}

func setupRoutes(e *echo.Echo, hub *ws.Hub) {
	e.Static("/static", "static")
	handlers.SetupHomeRoutes(e.Group(""), hub)
	handlers.SetupGameRoutes(e.Group("/game"), hub)
}
