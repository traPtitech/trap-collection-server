package migrate

import (
	"database/sql"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// v2
// v2用のテーブル作成
// データの移行は行わない
func v2() *gormigrate.Migration {
	tables := []any{
		&GameVersionTable2V2{},
		&GameFileTable2V2{},
		&GameImageTable2V2{},
		&GameVideoTable2V2{},
		&EditionTableV2{},
		&ProductKeyTableV2{},
		&AccessTokenTableV2{},
	}

	return &gormigrate.Migration{
		ID: "2",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(tables...)
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(tables...)
		},
	}
}

type GameTable2V2 struct {
	ID                  uuid.UUID                   `gorm:"type:varchar(36);not null;primaryKey"`
	Name                string                      `gorm:"type:varchar(256);size:256;not null"`
	Description         string                      `gorm:"type:text;not null"`
	CreatedAt           time.Time                   `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt              `gorm:"type:DATETIME NULL;default:NULL"`
	GameVersions        []GameVersionTable2V2       `gorm:"foreignkey:GameID"`
	GameManagementRoles []gameManagementRoleTableV1 `gorm:"foreignKey:GameID"`
	GameFiles           []GameFileTable2V2          `gorm:"foreignKey:GameID"`
	GameImages          []GameImageTable2V2         `gorm:"foreignKey:GameID"`
	GameVideos          []GameVideoTable2V2         `gorm:"foreignKey:GameID"`
}

func (*GameTable2V2) TableName() string {
	return "games"
}

type GameVersionTable2V2 struct {
	ID          uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	GameID      uuid.UUID `gorm:"type:varchar(36);not null"`
	GameImageID uuid.UUID `gorm:"type:varchar(36);not null"`
	GameVideoID uuid.UUID `gorm:"type:varchar(36);not null"`
	Name        string    `gorm:"type:varchar(32);size:32;not null"`
	Description string    `gorm:"type:text;not null"`
	CreatedAt   time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	// migrationのv2以降でも不自然でないように、
	// joinForeignKey、joinReferencesを指定している
	GameFiles []GameFileTable2V2 `gorm:"many2many:game_version_game_file_relations;joinForeignKey:GameVersionID;joinReferences:GameFileID"`
	GameImage GameImageTable2V2  `gorm:"foreignKey:GameImageID"`
	GameVideo GameVideoTable2V2  `gorm:"foreignKey:GameVideoID"`
}

func (*GameVersionTable2V2) TableName() string {
	return "v2_game_versions"
}

type GameFileTable2V2 struct {
	ID           uuid.UUID           `gorm:"type:varchar(36);not null;primaryKey"`
	GameID       uuid.UUID           `gorm:"type:varchar(36);not null"`
	FileTypeID   int                 `gorm:"type:tinyint;not null;index:idx_game_file_unique,unique"`
	Hash         string              `gorm:"type:char(32);size:32;not null"`
	EntryPoint   string              `gorm:"type:text;not null"`
	CreatedAt    time.Time           `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameFileType gameFileTypeTableV1 `gorm:"foreignKey:FileTypeID"`
}

func (*GameFileTable2V2) TableName() string {
	return "v2_game_files"
}

type GameImageTable2V2 struct {
	ID            uuid.UUID            `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID            `gorm:"type:varchar(36);not null"`
	ImageTypeID   int                  `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time            `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameImageType gameImageTypeTableV1 `gorm:"foreignKey:ImageTypeID"`
}

func (*GameImageTable2V2) TableName() string {
	return "v2_game_images"
}

type GameVideoTable2V2 struct {
	ID            uuid.UUID            `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID            `gorm:"type:varchar(36);not null"`
	VideoTypeID   int                  `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time            `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameVideoType gameVideoTypeTableV1 `gorm:"foreignKey:VideoTypeID"`
}

func (*GameVideoTable2V2) TableName() string {
	return "v2_game_videos"
}

type EditionTableV2 struct {
	ID               uuid.UUID             `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string                `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString        `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time             `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt        `gorm:"type:DATETIME NULL;default:NULL"`
	ProductKeys      []ProductKeyTableV2   `gorm:"foreignKey:EditionID"`
	GameVersions     []GameVersionTable2V2 `gorm:"many2many:edition_game_version_relations;joinForeignKey:EditionID;joinReferences:GameVersionID"`
}

func (*EditionTableV2) TableName() string {
	return "editions"
}

type ProductKeyTableV2 struct {
	ID           uuid.UUID            `gorm:"type:varchar(36);not null;primaryKey"`
	EditionID    uuid.UUID            `gorm:"type:varchar(36);not null"`
	ProductKey   string               `gorm:"type:varchar(29);not null;unique"`
	CreatedAt    time.Time            `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt    gorm.DeletedAt       `gorm:"type:DATETIME NULL;default:NULL"`
	AccessTokens []AccessTokenTableV2 `gorm:"foreignKey:ProductKeyID"`
}

func (*ProductKeyTableV2) TableName() string {
	return "product_keys"
}

type AccessTokenTableV2 struct {
	ID           uuid.UUID      `gorm:"type:varchar(36);not null;primaryKey"`
	ProductKeyID uuid.UUID      `gorm:"type:varchar(36);not null"`
	AccessToken  string         `gorm:"type:varchar(64);not null;unique"`
	ExpiresAt    time.Time      `gorm:"type:datetime;not null"`
	CreatedAt    time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt    gorm.DeletedAt `gorm:"type:DATETIME NULL;default:NULL"`
}

func (*AccessTokenTableV2) TableName() string {
	return "access_tokens"
}
