package main

import (
	"github.com/labstack/echo"
	align "github.com/nsip/curriculum-align"
	"log"
	"net/http"
)

func main() {
	align.Init()
	e := echo.New()
	e.GET("/align", align.Align)
	e.GET("/index", func(c echo.Context) error {
		query := c.QueryParam("search")
		ret, err := align.Search(query)
		if err != nil {
			return err
		} else {
			return c.String(http.StatusOK, string(ret))
		}
	})
	log.Println("Editor: localhost:1576")
	e.Logger.Fatal(e.Start(":1576"))
}
