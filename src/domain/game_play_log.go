package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GamePlayLog struct {
	ID            values.GamePlayLogID
	EditionID     values.LauncherVersionID
	GameID        values.GameID
	GameVersionID values.GameVersionID
	StartTime     time.Time
	EndTime       *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewGamePlayLog(
	editionID values.LauncherVersionID,
	gameID values.GameID,
	gameVersionID values.GameVersionID,
	startTime time.Time,
) *GamePlayLog {
	now := time.Now()
	return &GamePlayLog{
		ID:            values.NewGamePlayLogID(),
		EditionID:     editionID,
		GameID:        gameID,
		GameVersionID: gameVersionID,
		StartTime:     startTime,
		EndTime:       nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (g *GamePlayLog) GetID() values.GamePlayLogID {
	return g.ID
}

func (g *GamePlayLog) GetEditionID() values.LauncherVersionID {
	return g.EditionID
}

func (g *GamePlayLog) GetGameID() values.GameID {
	return g.GameID
}

func (g *GamePlayLog) GetGameVersionID() values.GameVersionID {
	return g.GameVersionID
}

func (g *GamePlayLog) GetStartTime() time.Time {
	return g.StartTime
}

func (g *GamePlayLog) GetEndTime() *time.Time {
	return g.EndTime
}

func (g *GamePlayLog) SetEndTime(endTime time.Time) {
	g.EndTime = &endTime
	g.UpdatedAt = time.Now()
}

func (g *GamePlayLog) GetCreatedAt() time.Time {
	return g.CreatedAt
}

func (g *GamePlayLog) GetUpdatedAt() time.Time {
	return g.UpdatedAt
}

func (g *GamePlayLog) GetPlayDuration() *time.Duration {
	if g.EndTime == nil {
		return nil
	}
	duration := g.EndTime.Sub(g.StartTime)
	return &duration
}

func (g *GamePlayLog) IsPlaying() bool {
	return g.EndTime == nil
}
