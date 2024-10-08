package streammanager

import (
	"github.com/prometheus/client_golang/prometheus"
	prom "github.com/zennittians/intelchain/api/service/prometheus"
)

func init() {
	prom.PromRegistry().MustRegister(
		discoverCounterVec,
		discoveredPeersCounterVec,
		addedStreamsCounterVec,
		removedStreamsCounterVec,
		setupStreamDuration,
		numStreamsGaugeVec,
	)
}

var (
	discoverCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "discover",
			Help:      "number of intentions to actively discover peers",
		},
		[]string{"topic"},
	)

	discoveredPeersCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "discover_peers",
			Help:      "number of peers discovered and connect actively",
		},
		[]string{"topic"},
	)

	addedStreamsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "added_streams",
			Help:      "number of streams added in stream manager",
		},
		[]string{"topic"},
	)

	removedStreamsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "removed_streams",
			Help:      "number of streams removed in stream manager",
		},
		[]string{"topic"},
	)

	setupStreamDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "setup_stream_duration",
			Help:      "duration in seconds of setting up connection to a discovered peer",
			// buckets: 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1280ms, +INF
			Buckets: prometheus.ExponentialBuckets(0.02, 2, 8),
		},
		[]string{"topic"},
	)

	numStreamsGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "num_streams",
			Help:      "number of connected streams",
		},
		[]string{"topic"},
	)
)
