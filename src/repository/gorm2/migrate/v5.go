package migrate

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// v5
// v2_game_filesのunique制約を解除
func v5() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "5",
		Migrate: func(tx *gorm.DB) error {
			err := tx.Migrator().DropConstraint(&gameFileTable2V5{}, "fk_v2_game_files_game_file_type")
			if err != nil {
				return fmt.Errorf("failed to drop constraint: %w", err)
			}

			err = tx.Migrator().DropIndex(&gameFileTable2V5{}, "idx_game_file_unique")
			if err != nil {
				return fmt.Errorf("failed to drop index: %w", err)
			}

			err = tx.AutoMigrate(&gameFileTable2V5{})
			if err != nil {
				return fmt.Errorf("failed to auto migrate: %w", err)
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&gameFileTable2V2{})
		},
	}
}

type gameFileTable2V5 struct {
	ID           uuid.UUID         `gorm:"type:varchar(36);not null;primaryKey"`
	GameID       uuid.UUID         `gorm:"type:varchar(36);not null"`
	FileTypeID   int               `gorm:"type:tinyint;not null"`
	Hash         string            `gorm:"type:char(32);size:32;not null"`
	EntryPoint   string            `gorm:"type:text;not null"`
	CreatedAt    time.Time         `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameFileType gameFileTypeTable `gorm:"foreignKey:FileTypeID"`
}

func (*gameFileTable2V5) TableName() string {
	return "v2_game_files"
}

type gameVersionTable2V5 struct {
	ID          uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	GameID      uuid.UUID `gorm:"type:varchar(36);not null"`
	GameImageID uuid.UUID `gorm:"type:varchar(36);not null"`
	GameVideoID uuid.UUID `gorm:"type:varchar(36);not null"`
	Name        string    `gorm:"type:varchar(32);size:32;not null"`
	Description string    `gorm:"type:text;not null"`
	URL         string    `gorm:"type:text;default:null"`
	CreatedAt   time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	// migrationのv2以降でも不自然でないように、
	// joinForeignKey、joinReferencesを指定している
	GameFiles []gameFileTable2V5 `gorm:"many2many:game_version_game_file_relations;joinForeignKey:GameVersionID;joinReferences:GameFileID"`
	GameImage gameImageTable2V2  `gorm:"foreignKey:GameImageID"`
	GameVideo gameVideoTable2V2  `gorm:"foreignKey:GameVideoID"`
}

func (*gameVersionTable2V5) TableName() string {
	return "v2_game_versions"
}

type editionTableV5 struct {
	ID               uuid.UUID             `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string                `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString        `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time             `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt        `gorm:"type:DATETIME NULL;default:NULL"`
	ProductKeys      []productKeyTableV2   `gorm:"foreignKey:EditionID"`
	GameVersions     []gameVersionTable2V5 `gorm:"many2many:edition_game_version_relations;joinForeignKey:EditionID;joinReferences:GameVersionID"`
}

func (*editionTableV5) TableName() string {
	return "editions"
}

type gameTable2V5 struct {
	ID                  uuid.UUID                 `gorm:"type:varchar(36);not null;primaryKey"`
	Name                string                    `gorm:"type:varchar(256);size:256;not null"`
	Description         string                    `gorm:"type:text;not null"`
	CreatedAt           time.Time                 `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt            `gorm:"type:DATETIME NULL;default:NULL"`
	GameVersions        []gameVersionTable2V5     `gorm:"foreignkey:GameID"`
	GameManagementRoles []gameManagementRoleTable `gorm:"foreignKey:GameID"`
	GameFiles           []gameFileTable2V5        `gorm:"foreignKey:GameID"`
	// GameImage2s
	// 不自然な名前だが、GameImagesだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameImage2s []gameImageTable2V2 `gorm:"foreignKey:GameID"`
	// GameVideo2s
	// 不自然な名前だが、GameVideosだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameVideo2s []gameVideoTable2V2 `gorm:"foreignKey:GameID"`
}

func (*gameTable2V5) TableName() string {
	return "games"
}
