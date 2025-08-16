package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type HourlyPlayStats struct {
	StartTime time.Time
	PlayCount int
	PlayTime  int // in seconds
}

type GamePlayStats struct {
	GameID           values.GameID
	TotalPlayCount   int
	TotalPlaySeconds int
	HourlyStats      []*HourlyPlayStats
}

type GamePlayStatsInEdition struct {
	GameID    values.GameID
	PlayCount int
	PlayTime  int // in seconds
}

type EditionPlayStats struct {
	EditionID        values.LauncherVersionID
	EditionName      values.LauncherVersionName
	TotalPlayCount   int
	TotalPlaySeconds int
	GameStats        []*GamePlayStatsInEdition
	HourlyStats      []*HourlyPlayStats
}
