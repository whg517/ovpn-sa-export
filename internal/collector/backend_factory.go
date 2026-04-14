package collector

import (
	"github.com/whg517/ovpn-sa-export/internal/backend/sacli"
	"github.com/whg517/ovpn-sa-export/internal/config"
)

func createBackend(cfg *config.Config) (Backend, error) {
	scfg := sacli.Config{
		Path:    cfg.Backend.Sacli.Path,
		Timeout: cfg.Backend.Sacli.Timeout,
	}
	return sacli.New(scfg), nil
}
