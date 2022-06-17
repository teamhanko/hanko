package main

import (
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/example/middleware"
	"log"
	"net/http"
)

func main() {
	e := echo.New()

	e.Use(mw.LoggerWithConfig(mw.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","time_unix":"${time_unix}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out},"referer":"${referer}"}` + "\n",
	}))

	e.Use(middleware.CacheControlMiddleware())

	e.File("/", "public/html/index.html")
	e.File("/secured", "public/html/secured.html", middleware.SessionMiddleware())
	e.File("/unauthorized", "public/html/unauthorized.html")

	e.GET("/logout", func(c echo.Context) error {
		cookie := &http.Cookie{
			Name:     "hanko",
			Value:    "",
			MaxAge:   -1,
			HttpOnly: true,
		}
		c.SetCookie(cookie)
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	})

	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
