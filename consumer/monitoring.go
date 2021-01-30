package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ConsumersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "consumer_consumer_count",
		Help: "The total number of running consumer",
	})
	MutatorsCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "consumer_mutators_count",
		Help: "The total number of running mutators",
	})
	ProducersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "consumer_producers_count",
		Help: "The total number of running producers",
	})
	InputStackSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "consumer_input_stack_size",
		Help: "The total number of items in the input stack",
	})
	OutputStackSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "consumer_output_stack_size",
		Help: "The total number of items in the output stack",
	})
)
