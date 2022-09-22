package migrate

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// v3
// v2でmigrationし忘れていたgameTable2のmigration
// 外部キー制約の追加のみ行われる
func v3() *gormigrate.Migration {
	tables := []any{
		&gameTable2V3{},
		&gameImageTable2V2{},
		&gameVideoTable2V2{},
	}

	return &gormigrate.Migration{
		ID: "3",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(tables...)
		},
		Rollback: func(tx *gorm.DB) error {
			err := tx.Migrator().DropConstraint(&gameImageTable2V2{}, "fk_games_game_image2s")
			if err != nil {
				return fmt.Errorf("failed to drop constraint fk_games_game_image2s: %w", err)
			}

			err = tx.Migrator().DropConstraint(&gameVideoTable2V2{}, "fk_games_game_video2s")
			if err != nil {
				return fmt.Errorf("failed to drop constraint fk_games_game_video2s: %w", err)
			}

			return nil
		},
	}
}

type gameTable2V3 struct {
	ID                  uuid.UUID                 `gorm:"type:varchar(36);not null;primaryKey"`
	Name                string                    `gorm:"type:varchar(256);size:256;not null"`
	Description         string                    `gorm:"type:text;not null"`
	CreatedAt           time.Time                 `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt            `gorm:"type:DATETIME NULL;default:NULL"`
	GameVersions        []gameVersionTable2V2     `gorm:"foreignkey:GameID"`
	GameManagementRoles []gameManagementRoleTable `gorm:"foreignKey:GameID"`
	GameFiles           []gameFileTable2V2        `gorm:"foreignKey:GameID"`
	// GameImage2s
	// 不自然な名前だが、GameImagesだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameImage2s []gameImageTable2V2 `gorm:"foreignKey:GameID"`
	// GameVideo2s
	// 不自然な名前だが、GameVideosだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameVideo2s []gameVideoTable2V2 `gorm:"foreignKey:GameID"`
}

func (*gameTable2V3) TableName() string {
	return "games"
}
