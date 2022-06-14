package middleware

import (
	"github.com/labstack/echo/v4"
)

func CacheControlMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			c.Response().Header().Set(echo.HeaderCacheControl, "no-store")

			return next(c)
		}
	}
}
