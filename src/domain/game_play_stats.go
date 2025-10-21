package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type HourlyPlayStats struct {
	startTime time.Time
	playCount int
	playTime  time.Duration
}

func (h *HourlyPlayStats) GetStartTime() time.Time {
	return h.startTime
}

func (h *HourlyPlayStats) GetPlayCount() int {
	return h.playCount
}

func (h *HourlyPlayStats) GetPlayTime() time.Duration {
	return h.playTime
}

func (h *HourlyPlayStats) AddPlayTime(d time.Duration) {
	h.playTime += d
}

func (h *HourlyPlayStats) AddPlayCount(n int) {
	h.playCount += n
}

type GamePlayStats struct {
	gameID         values.GameID
	totalPlayCount int
	totalPlayTime  time.Duration
	hourlyStats    []*HourlyPlayStats
}

func (g *GamePlayStats) GetGameID() values.GameID {
	return g.gameID
}

func (g *GamePlayStats) GetTotalPlayCount() int {
	return g.totalPlayCount
}

func (g *GamePlayStats) GetTotalPlayTime() time.Duration {
	return g.totalPlayTime
}

func (g *GamePlayStats) GetHourlyStats() []*HourlyPlayStats {
	return g.hourlyStats
}

type GamePlayStatsInEdition struct {
	gameID    values.GameID
	playCount int
	playTime  time.Duration
}

func (g *GamePlayStatsInEdition) GetGameID() values.GameID {
	return g.gameID
}

func (g *GamePlayStatsInEdition) GetPlayCount() int {
	return g.playCount
}

func (g *GamePlayStatsInEdition) GetPlayTime() time.Duration {
	return g.playTime
}

type EditionPlayStats struct {
	editionID      values.LauncherVersionID
	editionName    values.LauncherVersionName
	totalPlayCount int
	totalPlayTime  time.Duration
	gameStats      []*GamePlayStatsInEdition
	hourlyStats    []*HourlyPlayStats
}

func (e *EditionPlayStats) GetEditionID() values.LauncherVersionID {
	return e.editionID
}

func (e *EditionPlayStats) GetEditionName() values.LauncherVersionName {
	return e.editionName
}

func (e *EditionPlayStats) GetTotalPlayCount() int {
	return e.totalPlayCount
}

func (e *EditionPlayStats) GetTotalPlayTime() time.Duration {
	return e.totalPlayTime
}

func (e *EditionPlayStats) GetGameStats() []*GamePlayStatsInEdition {
	return e.gameStats
}

func (e *EditionPlayStats) GetHourlyStats() []*HourlyPlayStats {
	return e.hourlyStats
}

func NewHourlyPlayStats(startTime time.Time, playCount int, playTime time.Duration) *HourlyPlayStats {
	return &HourlyPlayStats{
		startTime: startTime,
		playCount: playCount,
		playTime:  playTime,
	}
}

func NewGamePlayStats(gameID values.GameID, totalPlayCount int, totalPlayTime time.Duration, hourlyStats []*HourlyPlayStats) *GamePlayStats {
	return &GamePlayStats{
		gameID:         gameID,
		totalPlayCount: totalPlayCount,
		totalPlayTime:  totalPlayTime,
		hourlyStats:    hourlyStats,
	}
}

func NewGamePlayStatsInEdition(gameID values.GameID, playCount int, playTime time.Duration) *GamePlayStatsInEdition {
	return &GamePlayStatsInEdition{
		gameID:    gameID,
		playCount: playCount,
		playTime:  playTime,
	}
}

func NewEditionPlayStats(editionID values.LauncherVersionID, editionName values.LauncherVersionName, totalPlayCount int, totalPlayTime time.Duration, gameStats []*GamePlayStatsInEdition, hourlyStats []*HourlyPlayStats) *EditionPlayStats {
	return &EditionPlayStats{
		editionID:      editionID,
		editionName:    editionName,
		totalPlayCount: totalPlayCount,
		totalPlayTime:  totalPlayTime,
		gameStats:      gameStats,
		hourlyStats:    hourlyStats,
	}
}
