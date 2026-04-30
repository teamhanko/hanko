package server

import (
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/handler"
	"github.com/teamhanko/hanko/backend/v2/mapper"
	"github.com/teamhanko/hanko/backend/v2/persistence"
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
	if !cfg.MultiTenancy {
		return
	}
	router := handler.NewManagementRouter(persister)
	router.Logger.Fatal(router.Start(cfg.Server.Management.Address))
}
