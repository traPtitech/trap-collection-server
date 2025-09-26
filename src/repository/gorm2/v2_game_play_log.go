package gorm2

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
)

type GamePlayLogV2 struct {
	db *DB
}

func NewGamePlayLogV2(db *DB) *GamePlayLogV2 {
	return &GamePlayLogV2{
		db: db,
	}
}

func (g *GamePlayLogV2) CreateGamePlayLog(ctx context.Context, playLog *domain.GamePlayLog) error {

	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var endTime sql.NullTime

	gamePlayLogTable := schema.GamePlayLogTable{
		ID:            uuid.UUID(playLog.GetID()),
		EditionID:     uuid.UUID(playLog.GetEditionID()),
		GameID:        uuid.UUID(playLog.GetGameID()),
		GameVersionID: uuid.UUID(playLog.GetGameVersionID()),
		StartTime:     playLog.GetStartTime(),
		EndTime:       endTime,
		CreatedAt:     playLog.GetCreatedAt(),
		UpdatedAt:     playLog.GetUpdatedAt(),
	}

	err = db.Create(&gamePlayLogTable).Error
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return repository.ErrDuplicatedUniqueKey
		}
		return fmt.Errorf("failed to create game play log: %w", err)
	}
	return nil
}

func (g *GamePlayLogV2) GetGamePlayLog(ctx context.Context, playLogID values.GamePlayLogID) (*domain.GamePlayLog, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}

func (g *GamePlayLogV2) UpdateGamePlayLogEndTime(ctx context.Context, playLogID values.GamePlayLogID, endTime time.Time) error {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}

func (g *GamePlayLogV2) GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}

func (g *GamePlayLogV2) GetEditionPlayStats(ctx context.Context, editionID values.LauncherVersionID, start, end time.Time) (*domain.EditionPlayStats, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}
