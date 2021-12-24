package swift

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

type GameImage struct {
	gameImageHitGauge *prometheus.GaugeVec
	client            *Client
}

func NewGameImage(client *Client) *GameImage {
	return &GameImage{
		gameImageHitGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "storage_trap_collection",
			Subsystem: "game_image",
			Name:      "cache_hit_count",
			Help:      "game image storage cache hit rate",
		}, []string{"result"}),
		client: client,
	}
}

func (gi *GameImage) SaveGameImage(ctx context.Context, reader io.Reader, imageID values.GameImageID) error {
	imageKey := gi.imageKey(imageID)

	err := gi.client.saveFile(
		ctx,
		imageKey,
		"",
		"",
		reader,
	)
	if errors.Is(err, ErrAlreadyExists) {
		return storage.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}

func (gi *GameImage) GetGameImage(ctx context.Context, writer io.Writer, image *domain.GameImage) error {
	imageKey := gi.imageKey(image.GetID())

	useCache, err := gi.client.loadFile(
		ctx,
		imageKey,
		writer,
	)
	if errors.Is(err, ErrNotFound) {
		return storage.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}

	if useCache {
		gi.gameImageHitGauge.
			WithLabelValues("hit").
			Inc()
	} else {
		gi.gameImageHitGauge.
			WithLabelValues("miss").
			Inc()
	}

	return nil
}

// imageKey 変更時にはオブジェクトストレージのキーを変更する必要があるので要注意
func (gi *GameImage) imageKey(imageID values.GameImageID) string {
	return fmt.Sprintf("images/%s", uuid.UUID(imageID).String())
}
