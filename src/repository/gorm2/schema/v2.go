package schema

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameTable2V15 struct {
	ID                     uuid.UUID                  `gorm:"type:varchar(36);not null;primaryKey"`
	Name                   string                     `gorm:"type:varchar(256);size:256;not null"`
	Description            string                     `gorm:"type:text;not null"`
	VisibilityTypeID       int                        `gorm:"type:tinyint;not null"`
	CreatedAt              time.Time                  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt              gorm.DeletedAt             `gorm:"type:DATETIME NULL;default:NULL"`
	LatestVersionUpdatedAt time.Time                  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameVersions           []GameVersionTable2V15     `gorm:"foreignkey:GameID"`
	GameManagementRoles    []GameManagementRoleTable  `gorm:"foreignKey:GameID"`
	GameVisibilityType     GameVisibilityTypeTableV11 `gorm:"foreignKey:VisibilityTypeID"`
	GameFiles              []GameFileTable2V5         `gorm:"foreignKey:GameID"`
	// GameImage2s
	// 不自然な名前だが、GameImagesだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameImage2s []GameImageTable2V2 `gorm:"foreignKey:GameID"`
	// GameVideo2s
	// 不自然な名前だが、GameVideosだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameVideo2s []GameVideoTable2V2 `gorm:"foreignKey:GameID"`
	// GameGenres
	// 後方参照を使っているためポインタになっている。
	// 参考: https://gorm.io/ja_JP/docs/many_to_many.html#%E5%BE%8C%E6%96%B9%E5%8F%82%E7%85%A7%EF%BC%88Back-Reference%EF%BC%89
	GameGenres []*GameGenreTableV15 `gorm:"many2many:game_genre_relations;joinForeignKey:GameID;joinReferences:GenreID"`
}

func (*GameTable2V15) TableName() string {
	return "games"
}

type GameVersionTable2V15 struct {
	ID          uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	GameID      uuid.UUID `gorm:"type:varchar(36);not null;uniqueIndex:idx_game_id_name"` // GameIDとNameの組み合わせでuniqueに
	GameImageID uuid.UUID `gorm:"type:varchar(36);not null"`
	GameVideoID uuid.UUID `gorm:"type:varchar(36);not null"`
	Name        string    `gorm:"type:varchar(32);size:32;not null;uniqueIndex:idx_game_id_name"` // GameIDとNameの組み合わせでuniqueに
	Description string    `gorm:"type:text;not null"`
	URL         string    `gorm:"type:text;default:null"`
	CreatedAt   time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	// migrationのv2以降でも不自然でないように、
	// joinForeignKey、joinReferencesを指定している
	GameFiles []GameFileTable2V5 `gorm:"many2many:game_version_game_file_relations;joinForeignKey:GameVersionID;joinReferences:GameFileID"`
	GameImage GameImageTable2V2  `gorm:"foreignKey:GameImageID"`
	GameVideo GameVideoTable2V2  `gorm:"foreignKey:GameVideoID"`
}

func (*GameVersionTable2V15) TableName() string {
	return "v2_game_versions"
}

type GameFileTable2V5 struct {
	ID           uuid.UUID         `gorm:"type:varchar(36);not null;primaryKey"`
	GameID       uuid.UUID         `gorm:"type:varchar(36);not null"`
	FileTypeID   int               `gorm:"type:tinyint;not null"`
	Hash         string            `gorm:"type:char(32);size:32;not null"`
	EntryPoint   string            `gorm:"type:text;not null"`
	CreatedAt    time.Time         `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameFileType GameFileTypeTable `gorm:"foreignKey:FileTypeID"`
}

func (*GameFileTable2V5) TableName() string {
	return "v2_game_files"
}

type GameImageTable2V2 struct {
	ID            uuid.UUID          `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID          `gorm:"type:varchar(36);not null"`
	ImageTypeID   int                `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time          `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameImageType GameImageTypeTable `gorm:"foreignKey:ImageTypeID"`
}

func (*GameImageTable2V2) TableName() string {
	return "v2_game_images"
}

type GameVideoTable2V2 struct {
	ID            uuid.UUID          `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID          `gorm:"type:varchar(36);not null"`
	VideoTypeID   int                `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time          `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameVideoType GameVideoTypeTable `gorm:"foreignKey:VideoTypeID"`
}

func (*GameVideoTable2V2) TableName() string {
	return "v2_game_videos"
}

type EditionTableV15 struct {
	ID               uuid.UUID              `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string                 `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString         `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time              `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt         `gorm:"type:DATETIME NULL;default:NULL"`
	ProductKeys      []ProductKeyTableV6    `gorm:"foreignKey:EditionID"`
	GameVersions     []GameVersionTable2V15 `gorm:"many2many:edition_game_version_relations;joinForeignKey:EditionID;joinReferences:GameVersionID"`
}

//nolint:unused
func (*EditionTableV15) TableName() string {
	return "editions"
}

type ProductKeyTableV6 struct {
	ID           uuid.UUID               `gorm:"type:varchar(36);not null;primaryKey"`
	EditionID    uuid.UUID               `gorm:"type:varchar(36);not null"`
	ProductKey   string                  `gorm:"type:varchar(29);not null;unique"`
	StatusID     int                     `gorm:"type:tinyint;not null"`
	CreatedAt    time.Time               `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	Status       ProductKeyStatusTableV6 `gorm:"foreignKey:StatusID"`
	AccessTokens []AccessTokenTableV2    `gorm:"foreignKey:ProductKeyID"`
}

func (*ProductKeyTableV6) TableName() string {
	return "product_keys"
}

type ProductKeyStatusTableV6 struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*ProductKeyStatusTableV6) TableName() string {
	return "product_key_statuses"
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

type AdminTable struct {
	UserID uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
}

func (*AdminTable) TableName() string {
	return "admins"
}

type SeatTableV9 struct {
	ID         uint              `gorm:"type:int;primaryKey;not null"`
	StatusID   uint8             `gorm:"type:tinyint;not null"`
	SeatStatus SeatStatusTableV9 `gorm:"foreignKey:StatusID"`
}

func (*SeatTableV9) TableName() string {
	return "seats"
}

type SeatStatusTableV9 struct {
	ID     uint8  `gorm:"type:tinyint;primaryKey;not null"`
	Name   string `gorm:"type:varchar(255);not null"`
	Active bool   `gorm:"type:boolean;not null;default:true"`
}

func (*SeatStatusTableV9) TableName() string {
	return "seat_statuses"
}

type GameGenreTableV15 struct {
	ID        uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	Name      string    `gorm:"type:varchar(32);not null;unique"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	// 後方参照を使っているためポインタになっている。
	// 参考: https://gorm.io/ja_JP/docs/many_to_many.html#%E5%BE%8C%E6%96%B9%E5%8F%82%E7%85%A7%EF%BC%88Back-Reference%EF%BC%89
	Games []*GameTable2V15 `gorm:"many2many:game_genre_relations;joinForeignKey:GenreID;joinReferences:GameID"`
}

func (*GameGenreTableV15) TableName() string {
	return "game_genres"
}

type GameVisibilityTypeTableV11 struct {
	ID        int       `gorm:"type:tinyint;not null;primaryKey"`
	Name      string    `gorm:"type:varchar(32);not null;unique"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

func (*GameVisibilityTypeTableV11) TableName() string {
	return "game_visibility_types"
}
