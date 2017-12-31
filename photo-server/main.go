package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())

	e.GET("/cameras", cameraIDRequestHandler())
	e.GET("/img/:camera_id", imgRequestHandler())

	e.Start(":8047")
}
