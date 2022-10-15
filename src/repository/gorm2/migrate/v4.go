package migrate

import (
	"database/sql"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// v4
// v2_game_versionsにurlカラムを追加
func v4() *gormigrate.Migration {
	tables := []any{
		&gameVersionTable2V4{},
	}

	return &gormigrate.Migration{
		ID: "4",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(tables...)
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropColumn(&gameVersionTable2V4{}, "url")
		},
	}
}

type gameVersionTable2V4 struct {
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
	GameFiles []gameFileTable2V2 `gorm:"many2many:game_version_game_file_relations;joinForeignKey:GameVersionID;joinReferences:GameFileID"`
	GameImage gameImageTable2V2  `gorm:"foreignKey:GameImageID"`
	GameVideo gameVideoTable2V2  `gorm:"foreignKey:GameVideoID"`
}

func (*gameVersionTable2V4) TableName() string {
	return "v2_game_versions"
}

//nolint:unused
type gameTable2V4 struct {
	ID                  uuid.UUID                 `gorm:"type:varchar(36);not null;primaryKey"`
	Name                string                    `gorm:"type:varchar(256);size:256;not null"`
	Description         string                    `gorm:"type:text;not null"`
	CreatedAt           time.Time                 `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt            `gorm:"type:DATETIME NULL;default:NULL"`
	GameVersions        []gameVersionTable2V4     `gorm:"foreignkey:GameID"`
	GameManagementRoles []gameManagementRoleTable `gorm:"foreignKey:GameID"`
	GameFiles           []gameFileTable2V2        `gorm:"foreignKey:GameID"`
	// GameImage2s
	// 不自然な名前だが、GameImagesだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameImage2s []gameImageTable2V2 `gorm:"foreignKey:GameID"`
	// GameVideo2s
	// 不自然な名前だが、GameVideosだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameVideo2s []gameVideoTable2V2 `gorm:"foreignKey:GameID"`
}

//nolint:unused
func (*gameTable2V4) TableName() string {
	return "games"
}

//nolint:unused
type editionTableV4 struct {
	ID               uuid.UUID             `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string                `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString        `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time             `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt        `gorm:"type:DATETIME NULL;default:NULL"`
	ProductKeys      []productKeyTableV2   `gorm:"foreignKey:EditionID"`
	GameVersions     []gameVersionTable2V4 `gorm:"many2many:edition_game_version_relations;joinForeignKey:EditionID;joinReferences:GameVersionID"`
}

//nolint:unused
func (*editionTableV4) TableName() string {
	return "editions"
}
