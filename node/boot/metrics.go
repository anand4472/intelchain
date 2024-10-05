package bootnode

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	prom "github.com/zennittians/intelchain/api/service/prometheus"
)

var (
	// nodeStringCounterVec is used to add version string or other static string
	// info into the metrics api
	nodeStringCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "bootnode",
			Name:      "metadata",
			Help:      "a list of boot node metadata",
		},
		[]string{"key", "value"},
	)

	onceMetrics sync.Once
)

func initMetrics() {
	onceMetrics.Do(func() {
		prom.PromRegistry().MustRegister(
			nodeStringCounterVec,
		)
	})
}
