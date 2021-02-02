package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SendCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "producer_input_send_count",
		Help: "The total number of items sent the input stack",
	})
)
