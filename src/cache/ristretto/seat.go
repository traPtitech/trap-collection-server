package ristretto

import (
	"context"
	"errors"
	"fmt"
	"time"

	ristretto "github.com/dgraph-io/ristretto/v2"
	"github.com/traPtitech/trap-collection-server/src/cache"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/domain"
)

type Seat struct {
	activeSeats    *ristretto.Cache[string, []*domain.Seat]
	activeSeatsTTL time.Duration
}

func NewSeat(conf config.CacheRistretto) (*Seat, error) {
	activeSeatsTTL, err := conf.ActiveSeatsTTL()
	if err != nil {
		return nil, fmt.Errorf("failed to get active seats ttl: %w", err)
	}

	activeSeats, err := ristretto.NewCache[string, []*domain.Seat](&ristretto.Config[string, []*domain.Seat]{
		NumCounters: 10,
		MaxCost:     64,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create activeUsers: %v", err)
	}

	return &Seat{
		activeSeats:    activeSeats,
		activeSeatsTTL: activeSeatsTTL,
	}, nil
}

func (seat *Seat) GetActiveSeats(_ context.Context) ([]*domain.Seat, error) {
	seats, ok := seat.activeSeats.Get(activeUsersKey)
	if !ok {
		hitCount.WithLabelValues("active_seats", "miss").Inc()
		return nil, cache.ErrCacheMiss
	}
	hitCount.WithLabelValues("active_seats", "hit").Inc()

	return seats, nil
}

func (seat *Seat) SetActiveSeats(_ context.Context, seats []*domain.Seat) error {
	// キャッシュ追加待ちのキューに入るだけで、すぐにはキャッシュが効かないのに注意
	ok := seat.activeSeats.SetWithTTL(
		activeUsersKey,
		seats,
		1,
		seat.activeSeatsTTL,
	)
	if !ok {
		return errors.New("failed to set activeUsers")
	}

	return nil
}
