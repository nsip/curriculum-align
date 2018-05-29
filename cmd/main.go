package main

import (
	"log"

	"github.com/labstack/echo"
	"github.com/nsip/curriculum-align"
)

func main() {
	align.Init()
	e := echo.New()
	e.GET("/align", align.Align)
	log.Println("Editor: localhost:1576")
	e.Logger.Fatal(e.Start(":1576"))
}
