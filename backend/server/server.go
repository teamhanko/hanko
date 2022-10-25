package server

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"sync"
)

func StartPublic(cfg *config.Config, wg *sync.WaitGroup, persister persistence.Persister) {
	defer wg.Done()
	router := NewPublicRouter(cfg, persister)
	router.Logger.Fatal(router.Start(cfg.Server.Public.Address))
}

func StartAdmin(cfg *config.Config, wg *sync.WaitGroup, persister persistence.Persister) {
	defer wg.Done()
	router := NewAdminRouter(persister)
	router.Logger.Fatal(router.Start(cfg.Server.Admin.Address))
}
