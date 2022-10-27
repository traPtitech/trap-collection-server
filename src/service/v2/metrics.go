package v2

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type SeatMetrics struct {
	Prefix      string
	activeGauge prometheus.Gauge
	emptyGauge  prometheus.Gauge
	inUseGauge  prometheus.Gauge
}

func NewSeatMetrics() *SeatMetrics {
	sm := &SeatMetrics{}

	sm.Prefix = "service_trap_collection"

	sm.activeGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: sm.Prefix,
		Subsystem: "seat_active",
		Name:      "count",
		Help:      "Number of active seats",
	})

	sm.emptyGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: sm.Prefix,
		Subsystem: "seat_empty",
		Name:      "count",
		Help:      "Number of empty seats",
	})

	sm.inUseGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: sm.Prefix,
		Subsystem: "seat_in_use",
		Name:      "count",
		Help:      "Number of seats in use",
	})

	return sm
}

func (sm *SeatMetrics) UpdateWithActiveSeats(activeSeats []*domain.Seat) {
	sm.activeGauge.Set(float64(len(activeSeats)))

	var emptyCount, inUseCount int
	for _, seat := range activeSeats {
		switch seat.Status() {
		case values.SeatStatusEmpty:
			emptyCount++
		case values.SeatStatusInUse:
			inUseCount++
		}
	}

	sm.emptyGauge.Set(float64(emptyCount))
	sm.inUseGauge.Set(float64(inUseCount))
}

func (sm *SeatMetrics) UpdateWithNewSeatStatus(newStatus values.SeatStatus) {
	switch newStatus {
	case values.SeatStatusEmpty:
		sm.emptyGauge.Inc()
		sm.inUseGauge.Dec()
	case values.SeatStatusInUse:
		sm.emptyGauge.Dec()
		sm.inUseGauge.Inc()
	}
}
