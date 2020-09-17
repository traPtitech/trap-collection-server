package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
)

// Maintainer gameのmaintainerの構造体
type Maintainer struct {
	ID        uint   `gorm:"type:int(11) unsigned auto_increment;PRIMARY_KEY;"`
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
	for _,user := range userIDs {
		interfaceUserIDs = append(interfaceUserIDs, Maintainer{
			GameID: gameID,
			UserID: user,
			Role: 0,
		})
	}

	err := gormbulk.BulkInsert(db, interfaceUserIDs, 3000, "CreatedAt", "DeletedAt")
	if err != nil {
		return fmt.Errorf("failed to bulk insert maintainers: %w", err)
	}

	return nil
}
