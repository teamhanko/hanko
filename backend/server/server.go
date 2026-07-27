package server

import (
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/handler"
	"github.com/teamhanko/hanko/backend/v3/mapper"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

func StartPublic(cfg *config.Config, wg *sync.WaitGroup, persister persistence.Persister, prometheus echo.MiddlewareFunc, authenticatorMetadata mapper.AuthenticatorMetadata) {
	defer wg.Done()
	router := handler.NewPublicRouter(cfg, persister, prometheus, authenticatorMetadata)
	router.Logger.Fatal(router.Start(cfg.Server.Public.Address))
}

func StartAdmin(cfg *config.Config, wg *sync.WaitGroup, persister persistence.Persister, prometheus echo.MiddlewareFunc) {
	defer wg.Done()
	router := handler.NewAdminRouter(cfg, persister, prometheus)
	router.Logger.Fatal(router.Start(cfg.Server.Admin.Address))
}

func StartManagement(cfg *config.Config, wg *sync.WaitGroup, persister persistence.Persister) {
	defer wg.Done()
	// DO not start the management server if multi-tenancy is disabled
	if cfg != nil && !cfg.MultiTenancy.Enabled {
		return
	}
	router := handler.NewManagementRouter(cfg, persister)
	router.Logger.Fatal(router.Start(cfg.Server.Management.Address))
}
