package collector

import (
	"github.com/whg517/openvpn-as-exporter/internal/backend/sacli"
	"github.com/whg517/openvpn-as-exporter/internal/config"
)

func createBackend(cfg *config.Config) (Backend, error) {
	scfg := sacli.Config{
		Path:    cfg.Backend.Sacli.Path,
		Timeout: cfg.Backend.Sacli.Timeout,
	}
	return sacli.New(scfg), nil
}
