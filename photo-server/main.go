package main

import (
	"flag"
	"fmt"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	port *int
)

func init() {
	port = flag.Int("p", 3000, "port number for forwarding")
	flag.Parse()

}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())

	e.GET("/cameras", cameraIDRequestHandler())
	e.GET("/img/:camera_id", imgRequestHandler())

	e.Start(fmt.Sprintf(":%d", *port))
}
