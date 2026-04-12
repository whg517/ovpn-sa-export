package collector

import (
	"github.com/whg517/ovpn-sa-export/internal/backend/sacli"
	"github.com/whg517/ovpn-sa-export/internal/backend/xmlrpc"
	"github.com/whg517/ovpn-sa-export/internal/config"
)

func createBackend(cfg *config.Config) (Backend, error) {
	switch cfg.Backend.Mode {
	case "xmlrpc":
		xcfg := xmlrpc.Config{
			Endpoint:           cfg.Backend.XMLRPC.Endpoint,
			Username:           cfg.Backend.XMLRPC.Username,
			Password:           cfg.Backend.XMLRPC.Password,
			SocketPath:         cfg.Backend.XMLRPC.SocketPath,
			Timeout:            cfg.Backend.XMLRPC.Timeout,
			InsecureSkipVerify: cfg.Backend.XMLRPC.InsecureSkipVerify,
		}
		return xmlrpc.New(xcfg)
	default:
		// Default to sacli
		scfg := sacli.Config{
			Path:    cfg.Backend.Sacli.Path,
			Timeout: cfg.Backend.Sacli.Timeout,
		}
		return sacli.New(scfg), nil
	}
}
