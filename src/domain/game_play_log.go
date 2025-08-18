package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GamePlayLog struct {
	id            values.GamePlayLogID
	editionID     values.LauncherVersionID
	gameID        values.GameID
	gameVersionID values.GameVersionID
	startTime     time.Time
	endTime       *time.Time
	createdAt     time.Time
	updatedAt     time.Time
}

func NewGamePlayLog(
	editionID values.LauncherVersionID,
	gameID values.GameID,
	gameVersionID values.GameVersionID,
	startTime time.Time,
) *GamePlayLog {
	now := time.Now()
	return &GamePlayLog{
		id:            values.NewGamePlayLogID(),
		editionID:     editionID,
		gameID:        gameID,
		gameVersionID: gameVersionID,
		startTime:     startTime,
		endTime:       nil,
		createdAt:     now,
		updatedAt:     now,
	}
}

func (g *GamePlayLog) GetID() values.GamePlayLogID {
	return g.id
}

func (g *GamePlayLog) GetEditionID() values.LauncherVersionID {
	return g.editionID
}

func (g *GamePlayLog) GetGameID() values.GameID {
	return g.gameID
}

func (g *GamePlayLog) GetGameVersionID() values.GameVersionID {
	return g.gameVersionID
}

func (g *GamePlayLog) GetStartTime() time.Time {
	return g.startTime
}

func (g *GamePlayLog) GetEndTime() *time.Time {
	return g.endTime
}

func (g *GamePlayLog) SetEndTime(endTime time.Time) {
	g.endTime = &endTime
	g.updatedAt = time.Now()
}

func (g *GamePlayLog) GetCreatedAt() time.Time {
	return g.createdAt
}

func (g *GamePlayLog) GetUpdatedAt() time.Time {
	return g.updatedAt
}

func (g *GamePlayLog) GetPlayDuration() *time.Duration {
	if g.endTime == nil {
		return nil
	}
	duration := g.endTime.Sub(g.startTime)
	return &duration
}

func (g *GamePlayLog) IsPlaying() bool {
	return g.endTime == nil
}
