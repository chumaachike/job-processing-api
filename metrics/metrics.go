package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	JobCreated prometheus.Counter
}

func New() *Metrics {
	m := &Metrics{
		JobCreated: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "jobs_created_total",
				Help: "Total number of jobs created",
			},
		),
	}

	prometheus.MustRegister(m.JobCreated)

	return m
}
