package server

import (
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"sync"
)

func StartPublic(cfg *config.Config, wg *sync.WaitGroup, persister persistence.Persister, prometheus *prometheus.Prometheus) {
	defer wg.Done()
	router := NewPublicRouter(cfg, persister, prometheus)
	router.Logger.Fatal(router.Start(cfg.Server.Public.Address))
}

func StartAdmin(cfg *config.Config, wg *sync.WaitGroup, persister persistence.Persister, prometheus *prometheus.Prometheus) {
	defer wg.Done()
	router := NewAdminRouter(cfg, persister, prometheus)
	router.Logger.Fatal(router.Start(cfg.Server.Admin.Address))
}
