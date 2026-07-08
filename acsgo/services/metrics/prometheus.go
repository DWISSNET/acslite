// Package metrics exposes Prometheus metrics for ACSGO.
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collector holds all ACSGO Prometheus metrics.
type Collector struct {
	informsTotal      prometheus.Counter
	deviceOnline      prometheus.Gauge
	deviceOffline     prometheus.Gauge
	cwmpLatency       prometheus.Histogram
	commandsTotal     *prometheus.CounterVec
	commandLatency    prometheus.Histogram
	rpcErrors         prometheus.Counter
	activeSessions    prometheus.Gauge
}

// NewCollector creates and registers all Prometheus metrics.
func NewCollector() *Collector {
	return &Collector{
		informsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "acsgo",
			Name:      "informs_total",
			Help:      "Total number of TR-069 Inform messages received.",
		}),
		deviceOnline: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "acsgo",
			Name:      "devices_online",
			Help:      "Number of currently online devices.",
		}),
		deviceOffline: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "acsgo",
			Name:      "devices_offline",
			Help:      "Number of currently offline devices.",
		}),
		cwmpLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "acsgo",
			Name:      "cwmp_request_duration_seconds",
			Help:      "CWMP request processing duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		}),
		commandsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "acsgo",
			Name:      "commands_total",
			Help:      "Total RPC commands by type and status.",
		}, []string{"type", "status"}),
		commandLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "acsgo",
			Name:      "command_delivery_duration_seconds",
			Help:      "RPC command delivery duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		}),
		rpcErrors: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "acsgo",
			Name:      "rpc_errors_total",
			Help:      "Total number of RPC delivery errors.",
		}),
		activeSessions: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "acsgo",
			Name:      "active_sessions",
			Help:      "Number of active CWMP sessions.",
		}),
	}
}

// RecordInform increments the inform counter.
func (c *Collector) RecordInform() {
	c.informsTotal.Inc()
}

// RecordCWMPRequest records CWMP request duration.
func (c *Collector) RecordCWMPRequest(d time.Duration) {
	c.cwmpLatency.Observe(d.Seconds())
}

// RecordCommand records a command event.
func (c *Collector) RecordCommand(cmdType, status string) {
	c.commandsTotal.WithLabelValues(cmdType, status).Inc()
}

// RecordCommandLatency records time to deliver a command.
func (c *Collector) RecordCommandLatency(d time.Duration) {
	c.commandLatency.Observe(d.Seconds())
}

// RecordRPCError increments the RPC error counter.
func (c *Collector) RecordRPCError() {
	c.rpcErrors.Inc()
}

// SetDeviceCounts updates the online/offline gauges.
func (c *Collector) SetDeviceCounts(online, offline int) {
	c.deviceOnline.Set(float64(online))
	c.deviceOffline.Set(float64(offline))
}

// SessionOpened increments the active session gauge.
func (c *Collector) SessionOpened() {
	c.activeSessions.Inc()
}

// SessionClosed decrements the active session gauge.
func (c *Collector) SessionClosed() {
	c.activeSessions.Dec()
}
