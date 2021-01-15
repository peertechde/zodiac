package server

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(sshConnsOpen)
	prometheus.MustRegister(sshConnsTotal)
	prometheus.MustRegister(sshConnsAttemptsTotal)
	prometheus.MustRegister(sshConnsAttemptsErrors)
	prometheus.MustRegister(sshReadCalls)
	prometheus.MustRegister(sshReadBytes)
	prometheus.MustRegister(sshReadErrors)
	prometheus.MustRegister(sshReadTimeout)
	prometheus.MustRegister(sshWriteCalls)
	prometheus.MustRegister(sshWriteBytes)
	prometheus.MustRegister(sshWriteErrors)
	prometheus.MustRegister(sshWriteTimeout)
}

var (
	sshConnsOpen = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ssh",
		Name:      "connections_open",
	})
	sshConnsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ssh",
		Name:      "connections_total",
	})
	sshConnsAttemptsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ssh",
		Name:      "connections_attempts_total",
	})
	sshConnsAttemptsErrors = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ssh",
		Name:      "connections_attempts_errors_total",
	})
	sshReadCalls = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ssh",
		Name:      "read_calls_total",
	})
	sshReadBytes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ssh",
		Name:      "read_bytes_total",
	})
	sshReadErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ssh",
		Name:      "read_errors_total",
	})
	sshReadTimeout = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ssh",
		Name:      "read_timeouts_total",
	})
	sshWriteCalls = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ssh",
		Name:      "write_calls_total",
	})
	sshWriteBytes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ssh",
		Name:      "write_bytes_total",
	})
	sshWriteErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ssh",
		Name:      "write_errors_total",
	})
	sshWriteTimeout = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ssh",
		Name:      "write_timeouts_total",
	})
)
