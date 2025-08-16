package service

import (
	"context"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type GamePlayLogV2 interface {
	CreatePlayLog(ctx context.Context, editionID values.LauncherVersionID, gameID values.GameID, gameVersionID values.GameVersionID, startTime time.Time) (*domain.GamePlayLog, error)
	UpdatePlayLogEndTime(ctx context.Context, playLogID values.GamePlayLogID, endTime time.Time) error
	GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error)
	GetEditionPlayStats(ctx context.Context, editionID values.LauncherVersionID, start, end time.Time) (*domain.EditionPlayStats, error)
}
