package migrate

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gameVisibilityTypeTableV11 struct {
	ID        int       `gorm:"type:tinyint;not null;primaryKey"`
	Name      string    `gorm:"type:varchar(32);not null;unique"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

func (*gameVisibilityTypeTableV11) TableName() string {
	return "game_visibility_types"
}

type gameTable2V11 struct {
	ID                  uuid.UUID                  `gorm:"type:varchar(36);not null;primaryKey"`
	Name                string                     `gorm:"type:varchar(256);size:256;not null"`
	Description         string                     `gorm:"type:text;not null"`
	VisibilityTypeID    int                        `gorm:"type:tinyint;not null"`
	CreatedAt           time.Time                  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt             `gorm:"type:DATETIME NULL;default:NULL"`
	GameVersions        []gameVersionTable2V5      `gorm:"foreignKey:GameID"`
	GameManagementRoles []gameManagementRoleTable  `gorm:"foreignKey:GameID"`
	GameVisibilityType  gameVisibilityTypeTableV11 `gorm:"foreignKey:VisibilityTypeID"`
	GameFiles           []gameFileTable2V5         `gorm:"foreignKey:GameID"`
	// GameImage2s
	// 不自然な名前だが、GameImagesだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameImage2s []gameImageTable2V2 `gorm:"foreignKey:GameID"`
	// GameVideo2s
	// 不自然な名前だが、GameVideosだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameVideo2s []gameVideoTable2V2 `gorm:"foreignKey:GameID"`
}

func (*gameTable2V11) TableName() string {
	return "games"
}

type gameGenreTableV11 struct {
	ID        uuid.UUID       `gorm:"type:varchar(36);not null;primaryKey"`
	Name      string          `gorm:"type:varchar(32);not null;unique"`
	CreatedAt time.Time       `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	Games     []gameTable2V11 `gorm:"many2many:game_genre_relations;joinForeignKey:GenreID;joinReferences:GameID"`
}

func (*gameGenreTableV11) TableName() string {
	return "game_genres"
}

func v11() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "11",
		Migrate: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&gameVisibilityTypeTableV11{})
			if err != nil {
				return fmt.Errorf("failed to migrate game visibility type table: %w", err)
			}

			err = setUpGameVisibilityTypeTable(tx)
			if err != nil {
				return fmt.Errorf("failed to set up game visibility type table: %w", err)
			}

			var visibilityPrivate gameVisibilityTypeTableV11
			err = tx.
				Where("name = ?", gameVisibilityTypePrivateV11).
				Find(&visibilityPrivate).
				Error
			if err != nil {
				return fmt.Errorf("failed to get game visibility type: %w", err)
			}

			err = tx.
				Exec("ALTER TABLE `games` ADD COLUMN `visibility_type_id` tinyint(4) NOT NULL DEFAULT ?", visibilityPrivate.ID).
				Error
			if err != nil {
				return fmt.Errorf("failed to add column on games table: %w", err)
			}

			err = tx.AutoMigrate(&gameTable2V11{}, &gameGenreTableV11{})
			if err != nil {
				return fmt.Errorf("failed to migrate games table: %w", err)
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&gameTable2V5{})
			if err != nil {
				return fmt.Errorf("failed to rollback games table: %w", err)
			}

			err = tx.Migrator().DropColumn(&gameTable2V11{}, "visibility_type_id")
			if err != nil {
				return fmt.Errorf("failed to drop visibility type id column on games table: %w", err)
			}

			err = tx.Migrator().DropTable(&gameVisibilityTypeTableV11{})
			if err != nil {
				return fmt.Errorf("failed to drop game visibility type table: %w", err)
			}

			return nil
		},
	}
}

const (
	gameVisibilityTypePublicV11  = "public"
	gameVisibilityTypeLimitedV11 = "limited"
	gameVisibilityTypePrivateV11 = "private"
)

func setUpGameVisibilityTypeTable(db *gorm.DB) error {
	visibilityList := []gameVisibilityTypeTableV11{
		{Name: gameVisibilityTypePublicV11},
		{Name: gameVisibilityTypeLimitedV11},
		{Name: gameVisibilityTypePrivateV11},
	}

	for _, visibility := range visibilityList {
		err := db.
			Session(&gorm.Session{}).
			Where("name = ?", visibility.Name).
			FirstOrCreate(&visibility).
			Error
		if err != nil {
			return fmt.Errorf("failed to create game visibility type: %w", err)
		}
	}

	return nil
}
