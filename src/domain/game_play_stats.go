package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type HourlyPlayStats struct {
	StartTime time.Time
	PlayCount int
	PlayTime  time.Duration
}

type GamePlayStats struct {
	GameID         values.GameID
	TotalPlayCount int
	TotalPlayTime  time.Duration
	HourlyStats    []*HourlyPlayStats
}

type GamePlayStatsInEdition struct {
	GameID    values.GameID
	PlayCount int
	PlayTime  time.Duration
}

type EditionPlayStats struct {
	EditionID      values.LauncherVersionID
	EditionName    values.LauncherVersionName
	TotalPlayCount int
	TotalPlayTime  time.Duration
	GameStats      []*GamePlayStatsInEdition
	HourlyStats    []*HourlyPlayStats
}
