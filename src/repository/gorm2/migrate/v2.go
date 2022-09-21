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
		&gameVersionTable2V2{},
		&gameFileTable2V2{},
		&gameImageTable2V2{},
		&gameVideoTable2V2{},
		&editionTableV2{},
		&productKeyTableV2{},
		&accessTokenTableV2{},
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

type gameTable2V2 struct {
	ID                  uuid.UUID                 `gorm:"type:varchar(36);not null;primaryKey"`
	Name                string                    `gorm:"type:varchar(256);size:256;not null"`
	Description         string                    `gorm:"type:text;not null"`
	CreatedAt           time.Time                 `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt            `gorm:"type:DATETIME NULL;default:NULL"`
	GameVersions        []gameVersionTable2V2     `gorm:"foreignkey:GameID"`
	GameManagementRoles []gameManagementRoleTable `gorm:"foreignKey:GameID"`
	GameFiles           []gameFileTable2V2        `gorm:"foreignKey:GameID"`
	GameImages          []gameImageTable2V2       `gorm:"foreignKey:GameID"`
	GameVideos          []gameVideoTable2V2       `gorm:"foreignKey:GameID"`
}

func (*gameTable2V2) TableName() string {
	return "games"
}

type gameVersionTable2V2 struct {
	ID          uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	GameID      uuid.UUID `gorm:"type:varchar(36);not null"`
	GameImageID uuid.UUID `gorm:"type:varchar(36);not null"`
	GameVideoID uuid.UUID `gorm:"type:varchar(36);not null"`
	Name        string    `gorm:"type:varchar(32);size:32;not null"`
	Description string    `gorm:"type:text;not null"`
	CreatedAt   time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	// migrationのv2以降でも不自然でないように、
	// joinForeignKey、joinReferencesを指定している
	GameFiles []gameFileTable2V2 `gorm:"many2many:game_version_game_file_relations;joinForeignKey:GameVersionID;joinReferences:GameFileID"`
	GameImage gameImageTable2V2  `gorm:"foreignKey:GameImageID"`
	GameVideo gameVideoTable2V2  `gorm:"foreignKey:GameVideoID"`
}

func (*gameVersionTable2V2) TableName() string {
	return "v2_game_versions"
}

type gameFileTable2V2 struct {
	ID           uuid.UUID         `gorm:"type:varchar(36);not null;primaryKey"`
	GameID       uuid.UUID         `gorm:"type:varchar(36);not null"`
	FileTypeID   int               `gorm:"type:tinyint;not null;index:idx_game_file_unique,unique"`
	Hash         string            `gorm:"type:char(32);size:32;not null"`
	EntryPoint   string            `gorm:"type:text;not null"`
	CreatedAt    time.Time         `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameFileType gameFileTypeTable `gorm:"foreignKey:FileTypeID"`
}

func (*gameFileTable2V2) TableName() string {
	return "v2_game_files"
}

type gameImageTable2V2 struct {
	ID            uuid.UUID          `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID          `gorm:"type:varchar(36);not null"`
	ImageTypeID   int                `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time          `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameImageType gameImageTypeTable `gorm:"foreignKey:ImageTypeID"`
}

func (*gameImageTable2V2) TableName() string {
	return "v2_game_images"
}

type gameVideoTable2V2 struct {
	ID            uuid.UUID          `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID          `gorm:"type:varchar(36);not null"`
	VideoTypeID   int                `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time          `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameVideoType gameVideoTypeTable `gorm:"foreignKey:VideoTypeID"`
}

func (*gameVideoTable2V2) TableName() string {
	return "v2_game_videos"
}

type editionTableV2 struct {
	ID               uuid.UUID             `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string                `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString        `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time             `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt        `gorm:"type:DATETIME NULL;default:NULL"`
	ProductKeys      []productKeyTableV2   `gorm:"foreignKey:EditionID"`
	GameVersions     []gameVersionTable2V2 `gorm:"many2many:edition_game_version_relations;joinForeignKey:EditionID;joinReferences:GameVersionID"`
}

func (*editionTableV2) TableName() string {
	return "editions"
}

type productKeyTableV2 struct {
	ID           uuid.UUID            `gorm:"type:varchar(36);not null;primaryKey"`
	EditionID    uuid.UUID            `gorm:"type:varchar(36);not null"`
	ProductKey   string               `gorm:"type:varchar(29);not null;unique"`
	CreatedAt    time.Time            `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt    gorm.DeletedAt       `gorm:"type:DATETIME NULL;default:NULL"`
	AccessTokens []accessTokenTableV2 `gorm:"foreignKey:ProductKeyID"`
}

func (*productKeyTableV2) TableName() string {
	return "product_keys"
}

type accessTokenTableV2 struct {
	ID           uuid.UUID      `gorm:"type:varchar(36);not null;primaryKey"`
	ProductKeyID uuid.UUID      `gorm:"type:varchar(36);not null"`
	AccessToken  string         `gorm:"type:varchar(64);not null;unique"`
	ExpiresAt    time.Time      `gorm:"type:datetime;not null"`
	CreatedAt    time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt    gorm.DeletedAt `gorm:"type:DATETIME NULL;default:NULL"`
}

func (*accessTokenTableV2) TableName() string {
	return "access_tokens"
}
