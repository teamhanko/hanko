package main

import (
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/quickstart/middleware"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	t := &Template{
		templates: template.Must(template.ParseGlob("public/html/*.html")),
	}

	hankoUrl := getEnv("HANKO_URL")
	hankoElementUrl := getEnv("HANKO_ELEMENT_URL")
	hankoFrontendSdkUrl := getEnv("HANKO_FRONTEND_SDK_URL")
	hankoUrlInternal := hankoUrl
	domain := ""
	if value, ok := os.LookupEnv("HANKO_URL_INTERNAL"); ok {
		hankoUrlInternal = value
	}
	if value, ok := os.LookupEnv("DOMAIN"); ok {
		domain = value
	}

	e := echo.New()
	e.Renderer = t

	e.Use(mw.LoggerWithConfig(mw.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","time_unix":"${time_unix}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out},"referer":"${referer}"}` + "\n",
	}))

	e.Use(middleware.CacheControlMiddleware())

	e.Static("/static", "public/assets")

	e.GET("/", func(c echo.Context) error {
		indexData := IndexData{
			HankoUrl:        hankoUrl,
			HankoElementUrl: hankoElementUrl,
		}
		return c.Render(http.StatusOK, "index.html", &indexData)
	})

	e.File("/unauthorized", "public/html/unauthorized.html")
	e.File("/error", "public/html/error.html")

	e.GET("/secured", func(c echo.Context) error {
		return c.Render(http.StatusOK, "secured.html", map[string]interface{}{
			"HankoFrontendSdkUrl": hankoFrontendSdkUrl,
			"HankoUrl":            hankoUrl,
		})
	}, middleware.SessionMiddleware(hankoUrlInternal))

	e.GET("/logout", func(c echo.Context) error {
		cookie := &http.Cookie{
			Name:     "hanko",
			Value:    "",
			MaxAge:   -1,
			HttpOnly: true,
			Domain:   domain,
		}
		c.SetCookie(cookie)
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	})

	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}

type IndexData struct {
	HankoUrl        string
	HankoElementUrl string
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func getEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Fatalf("env key not set: %v", key)
	return ""
}
