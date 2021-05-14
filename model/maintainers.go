package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// Maintainer gameのmaintainerの構造体
type Maintainer struct {
	ID          string      `gorm:"type:varchar(36);PRIMARY_KEY;"`
	GameID    string `gorm:"type:varchar(36);NOT NULL;"`
	Game      Game
	UserID    string    `gorm:"type:varchar(36);NOT NULL;"`
	Role      uint8     `gorm:"type:tinyint;NOT NULL;DEFAULT:0;"`
	CreatedAt time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt time.Time `gorm:"type:datetime;DEFAULT:NULL;"`
}

// MaintainerMeta maintainerテーブルのリポジトリ
type MaintainerMeta interface {
	CheckMaintainerID(userID string, gameID string) (bool, error)
	InsertMaintainer(gameID string, userIDs []string) error
	GetMaintainers(gameID string, userMap map[string]*openapi.User) ([]*openapi.Maintainer, error)
}

// CheckMaintainerID ゲームの管理者のチェック
func (*DB) CheckMaintainerID(userID string, gameID string) (bool, error) {
	var maintainer Maintainer

	err := db.Select("user_id").
		Where("game_id = ? AND user_id = ?", gameID, userID).
		First(&maintainer).Error
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// InsertMaintainer 管理者の追加
func (*DB) InsertMaintainer(gameID string, userIDs []string) error {
	interfaceUserIDs := make([]interface{}, 0, len(userIDs))
	for _, user := range userIDs {
		interfaceUserIDs = append(interfaceUserIDs, Maintainer{
			ID: uuid.New().String(),
			GameID: gameID,
			UserID: user,
			Role:   0,
		})
	}

	err := gormbulk.BulkInsert(db, interfaceUserIDs, 3000, "CreatedAt", "DeletedAt")
	if err != nil {
		return fmt.Errorf("failed to bulk insert maintainers: %w", err)
	}

	return nil
}

// GetMaintainers ゲームの管理者の追加
func (*DB) GetMaintainers(gameID string, userMap map[string]*openapi.User) ([]*openapi.Maintainer, error) {
	maintainers := []Maintainer{}
	err := db.Where("game_id = ?", gameID).Find(&maintainers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get maintainers: %w", err)
	}

	apiMaintainers := make([]*openapi.Maintainer, 0, len(maintainers))
	for _, maintainer := range maintainers {
		user, ok := userMap[maintainer.UserID]
		if !ok {
			return nil, errors.New("invalid User error")
		}

		apiMaintainer := &openapi.Maintainer{
			Id:   maintainer.UserID,
			Name: user.Name,
			Role: int32(maintainer.Role),
		}

		apiMaintainers = append(apiMaintainers, apiMaintainer)
	}

	return apiMaintainers, nil
}
