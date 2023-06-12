package server

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/handler"
	"github.com/teamhanko/hanko/backend/persistence"
	"sync"
)

func StartPublic(cfg *config.Config, wg *sync.WaitGroup, persister persistence.Persister, prometheus echo.MiddlewareFunc) {
	defer wg.Done()
	router := handler.NewPublicRouter(cfg, persister, prometheus)
	router.Logger.Fatal(router.Start(cfg.Server.Public.Address))
}

func StartAdmin(cfg *config.Config, wg *sync.WaitGroup, persister persistence.Persister, prometheus echo.MiddlewareFunc) {
	defer wg.Done()
	router := handler.NewAdminRouter(cfg, persister, prometheus)
	router.Logger.Fatal(router.Start(cfg.Server.Admin.Address))
}
