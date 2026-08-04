package stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PacketsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gonetmon_packets_total",
		Help: "The total number of captured packets",
	}, []string{"interface", "protocol"})

	BytesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gonetmon_bytes_total",
		Help: "The total number of captured bytes",
	}, []string{"interface", "protocol"})

	PacketsPerSec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gonetmon_packets_per_sec",
		Help: "Current packets per second",
	}, []string{"interface"})

	BytesPerSec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gonetmon_bytes_per_sec",
		Help: "Current bytes per second",
	}, []string{"interface"})

	ActiveConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gonetmon_active_connections",
		Help: "Number of active connections in the tracking table",
	}, []string{"interface"})
)

// RecordPacket updates Prometheus metrics for a single packet.
func RecordPacket(iface, proto string, size int) {
	PacketsTotal.WithLabelValues(iface, proto).Inc()
	BytesTotal.WithLabelValues(iface, proto).Add(float64(size))
}

// UpdateRates updates Prometheus gauge metrics for rates.
func UpdateRates(iface string, pps, bps float64, connCount int) {
	PacketsPerSec.WithLabelValues(iface).Set(pps)
	BytesPerSec.WithLabelValues(iface).Set(bps)
	ActiveConnections.WithLabelValues(iface).Set(float64(connCount))
}
