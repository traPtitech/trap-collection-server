package ristretto

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var hitCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "cache_trap_collection",
	Name:      "hit_count",
}, []string{"type", "status"})
