package common

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	Abstracted_counter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "abstracted_metrics",
		Help: "metric that may be reported by server or plugin but not both",
	})
)

func RunAbstractedCounter(interval time.Duration, stepSize int) {
	prometheus.MustRegister(Abstracted_counter)

	t := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-t.C:

				Abstracted_counter.Add(float64(stepSize))
			}
		}
	}()
}
