package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/example/middleware"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	t := &Template{
		templates: template.Must(template.ParseGlob("public/html/*.html")),
	}

	hankoUrl := getEnv("HANKO_URL")
	hankoElementUrl := getEnv("HANKO_ELEMENT_URL")
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
	e.Renderer = t

	e.Static("/static", "public/assets")

	e.GET("/", func(c echo.Context) error {
		indexData := IndexData{
			HankoUrl:        hankoUrl,
			HankoElementUrl: hankoElementUrl,
		}
		return c.Render(http.StatusOK, "index.html", &indexData)
	})

	e.File("/unauthorized", "public/html/unauthorized.html")
	e.GET("/secured", func(c echo.Context) error {
		client := http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/%s", hankoUrlInternal, c.Get("user")), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Get("token")))
		if err != nil {
			return c.Render(http.StatusOK, "error.html", map[string]interface{}{
				"error": err.Error(),
			})
		}

		resp, err := client.Do(req)
		if err != nil {
			return c.Render(http.StatusOK, "error.html", map[string]interface{}{
				"error": err.Error(),
			})
		}

		if resp != nil {
			defer resp.Body.Close()
		}

		user := struct {
			ID                  string    `json:"id"`
			Email               string    `json:"email"`
			CreatedAt           time.Time `json:"created_at"`
			UpdatedAt           time.Time `json:"updated_at"`
			Verified            bool      `json:"verified"`
			WebauthnCredentials []struct {
				ID string `json:"id"`
			} `json:"webauthn_credentials"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&user)

		if err != nil {
			return c.Render(http.StatusOK, "error.html", map[string]interface{}{
				"error": err.Error(),
			})
		}

		return c.Render(http.StatusOK, "secured.html", map[string]interface{}{
			"user": user.Email,
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
