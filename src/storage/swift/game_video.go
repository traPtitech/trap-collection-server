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

type GameVideo struct {
	gameVideoHitGauge *prometheus.GaugeVec
	client            *Client
}

func NewGameVideo(client *Client) *GameVideo {
	return &GameVideo{
		gameVideoHitGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "storage_trap_collection",
			Subsystem: "game_video",
			Name:      "cache_hit_count",
			Help:      "game video storage cache hit rate",
		}, []string{"result"}),
		client: client,
	}
}

func (gv *GameVideo) SaveGameVideo(ctx context.Context, reader io.Reader, videoID values.GameVideoID) error {
	videoKey := gv.videoKey(videoID)

	err := gv.client.saveFile(
		ctx,
		videoKey,
		"",
		"",
		reader,
	)
	if errors.Is(err, ErrAlreadyExists) {
		return storage.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("failed to save video: %w", err)
	}

	return nil
}

func (gv *GameVideo) GetGameVideo(ctx context.Context, writer io.Writer, video *domain.GameVideo) error {
	videoKey := gv.videoKey(video.GetID())

	useCache, err := gv.client.loadFile(
		ctx,
		videoKey,
		writer,
	)
	if errors.Is(err, ErrNotFound) {
		return storage.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	if useCache {
		gv.gameVideoHitGauge.
			WithLabelValues("hit").
			Inc()
	} else {
		gv.gameVideoHitGauge.
			WithLabelValues("miss").
			Inc()
	}

	return nil
}

func (gv *GameVideo) videoKey(videoID values.GameVideoID) string {
	return fmt.Sprintf("videos/%s", uuid.UUID(videoID).String())
}
