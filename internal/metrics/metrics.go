package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/whg517/ovpn-sa-export/pkg/types"
)

// Registry holds all Prometheus metrics and a custom prometheus registry.
type Registry struct {
	promRegistry      *prometheus.Registry
	up                prometheus.Gauge
	connectedClients  prometheus.Gauge
	dcoAvailable      prometheus.Gauge
	dcoVersion        *prometheus.GaugeVec
	clientBytesRecv   *prometheus.GaugeVec
	clientBytesSent   *prometheus.GaugeVec
	clientConnected   *prometheus.GaugeVec
	subCurrent        prometheus.Gauge
	subMax            prometheus.Gauge
	subFallback       prometheus.Gauge
	subLastUpdate     prometheus.Gauge
	serviceUp         *prometheus.GaugeVec
	scrapeDuration    *prometheus.HistogramVec
	scrapeTotal       *prometheus.CounterVec
	scrapeErrors      *prometheus.CounterVec
}

// NewRegistry creates and registers all metrics in a custom registry.
func NewRegistry() *Registry {
	r := &Registry{
		promRegistry: prometheus.NewRegistry(),
	}

	r.up = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_up",
		Help: "Whether the last scrape of OpenVPN AS was successful.",
	})
	r.connectedClients = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_connected_clients",
		Help: "Number of currently connected VPN clients.",
	})
	r.dcoAvailable = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_dco_available",
		Help: "Whether OpenVPN DCO (Data Channel Offload) is available.",
	})
	r.dcoVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_dco_version",
		Help: "OpenVPN DCO version info (1 if available).",
	}, []string{"version"})

	r.clientBytesRecv = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_client_bytes_received",
		Help: "Total bytes received from a VPN client.",
	}, []string{"username", "common_name", "real_addr", "virtual_addr"})
	r.clientBytesSent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_client_bytes_sent",
		Help: "Total bytes sent to a VPN client.",
	}, []string{"username", "common_name", "real_addr", "virtual_addr"})
	r.clientConnected = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_client_connected_since",
		Help: "UNIX timestamp when the VPN client connected.",
	}, []string{"username", "common_name"})

	r.subCurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_subscription_current_connections",
		Help: "Current number of client connections in use.",
	})
	r.subMax = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_subscription_max_connections",
		Help: "Maximum number of client connections allowed by subscription.",
	})
	r.subFallback = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_subscription_fallback_connections",
		Help: "Number of fallback connections in use.",
	})
	r.subLastUpdate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_subscription_last_update_timestamp",
		Help: "UNIX timestamp of the last successful subscription update.",
	})

	r.serviceUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ovpn_sa_export_service_up",
		Help: "Whether an internal AS service is running.",
	}, []string{"service"})

	r.scrapeDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ovpn_sa_export_scrape_duration_seconds",
		Help:    "Duration of the last scrape for each collector.",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"collector"})
	r.scrapeTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ovpn_sa_export_scrapes_total",
		Help: "Total number of scrapes.",
	}, []string{"collector"})
	r.scrapeErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ovpn_sa_export_scrape_errors_total",
		Help: "Total number of scrape errors.",
	}, []string{"collector"})

	r.promRegistry.MustRegister(
		r.up,
		r.connectedClients,
		r.dcoAvailable,
		r.dcoVersion,
		r.clientBytesRecv,
		r.clientBytesSent,
		r.clientConnected,
		r.subCurrent,
		r.subMax,
		r.subFallback,
		r.subLastUpdate,
		r.serviceUp,
		r.scrapeDuration,
		r.scrapeTotal,
		r.scrapeErrors,
	)
	return r
}

// PromRegistry returns the underlying prometheus.Registry for use with promhttp.Handler.
func (r *Registry) PromRegistry() *prometheus.Registry {
	return r.promRegistry
}

// RecordScrape records a scrape attempt result.
func (r *Registry) RecordScrape(collector string, duration time.Duration, err error) {
	r.scrapeDuration.WithLabelValues(collector).Observe(duration.Seconds())
	r.scrapeTotal.WithLabelValues(collector).Inc()
	if err != nil {
		r.scrapeErrors.WithLabelValues(collector).Inc()
	}
}

// UpdateVPNStatus updates per-client VPN metrics.
func (r *Registry) UpdateVPNStatus(clients []types.VPNClientStatus) {
	r.up.Set(1)
	r.connectedClients.Set(float64(len(clients)))

	for _, c := range clients {
		labels := prometheus.Labels{
			"username":     c.Username,
			"common_name":  c.CommonName,
			"real_addr":    c.RealAddress,
			"virtual_addr": c.VirtualAddress,
		}
		r.clientBytesRecv.With(labels).Set(float64(c.BytesReceived))
		r.clientBytesSent.With(labels).Set(float64(c.BytesSent))
		r.clientConnected.With(prometheus.Labels{
			"username":    c.Username,
			"common_name": c.CommonName,
		}).Set(float64(c.ConnectedSince.Unix()))
	}
}

// UpdateVPNSummary updates summary metrics.
func (r *Registry) UpdateVPNSummary(s *types.VPNSummary) {
	if s == nil {
		return
	}
	r.connectedClients.Set(float64(s.NClients))
	if s.OvpnDcoAvailable {
		r.dcoAvailable.Set(1)
		r.dcoVersion.WithLabelValues(s.OvpnDcoVersion).Set(1)
	} else {
		r.dcoAvailable.Set(0)
	}
}

// UpdateSubscription updates subscription/license metrics.
func (r *Registry) UpdateSubscription(s *types.SubscriptionStatus) {
	if s == nil {
		return
	}
	r.subCurrent.Set(float64(s.CurrentConnections))
	r.subMax.Set(float64(s.MaxConnections))
	r.subFallback.Set(float64(s.FallbackConnections))
	r.subLastUpdate.Set(float64(s.LastSuccessfulUpdate.Unix()))
}

// UpdateServiceStatus updates service health metrics.
func (r *Registry) UpdateServiceStatus(s *types.ServiceStatus) {
	if s == nil {
		return
	}
	for name, running := range s.Services {
		val := float64(0)
		if running {
			val = 1
		}
		r.serviceUp.WithLabelValues(name).Set(val)
	}
}
