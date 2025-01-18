package migrate

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gameVersionTable2V15 struct {
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
	GameFiles []gameFileTable2V5 `gorm:"many2many:game_version_game_file_relations;joinForeignKey:GameVersionID;joinReferences:GameFileID"`
	GameImage gameImageTable2V2  `gorm:"foreignKey:GameImageID"`
	GameVideo gameVideoTable2V2  `gorm:"foreignKey:GameVideoID"`
}

func (*gameVersionTable2V15) TableName() string {
	return "v2_game_versions"
}

//nolint:unused
type gameTable2V15 struct {
	ID                     uuid.UUID                  `gorm:"type:varchar(36);not null;primaryKey"`
	Name                   string                     `gorm:"type:varchar(256);size:256;not null"`
	Description            string                     `gorm:"type:text;not null"`
	VisibilityTypeID       int                        `gorm:"type:tinyint;not null"`
	CreatedAt              time.Time                  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt              gorm.DeletedAt             `gorm:"type:DATETIME NULL;default:NULL"`
	LatestVersionUpdatedAt time.Time                  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameVersions           []gameVersionTable2V15     `gorm:"foreignKey:GameID"`
	GameManagementRoles    []gameManagementRoleTable  `gorm:"foreignKey:GameID"`
	GameVisibilityType     gameVisibilityTypeTableV11 `gorm:"foreignKey:VisibilityTypeID"`
	GameFiles              []gameFileTable2V5         `gorm:"foreignKey:GameID"`
	// GameImage2s
	// 不自然な名前だが、GameImagesだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameImage2s []gameImageTable2V2 `gorm:"foreignKey:GameID"`
	// GameVideo2s
	// 不自然な名前だが、GameVideosだとアプリケーションv1とforeign key名が被るためこの名前にしている
	GameVideo2s []gameVideoTable2V2 `gorm:"foreignKey:GameID"`
	// GameGenres
	// 後方参照を使っているためポインタになっている。
	// 参考: https://gorm.io/ja_JP/docs/many_to_many.html#%E5%BE%8C%E6%96%B9%E5%8F%82%E7%85%A7%EF%BC%88Back-Reference%EF%BC%89
	GameGenres []*gameGenreTableV15 `gorm:"many2many:game_genre_relations;joinForeignKey:GameID;joinReferences:GenreID"`
}

type gameGenreTableV15 struct {
	ID        uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	Name      string    `gorm:"type:varchar(32);not null;unique"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	// 後方参照を使っているためポインタになっている。
	// 参考: https://gorm.io/ja_JP/docs/many_to_many.html#%E5%BE%8C%E6%96%B9%E5%8F%82%E7%85%A7%EF%BC%88Back-Reference%EF%BC%89
	Games []*gameTable2V15 `gorm:"many2many:game_genre_relations;joinForeignKey:GenreID;joinReferences:GameID"`
}

//nolint:unused
func (*gameTable2V15) TableName() string {
	return "games"
}

func (*gameGenreTableV15) TableName() string {
	return "game_genres"
}

//nolint:unused
type editionTableV15 struct {
	ID               uuid.UUID              `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string                 `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString         `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time              `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt         `gorm:"type:DATETIME NULL;default:NULL"`
	ProductKeys      []productKeyTableV2    `gorm:"foreignKey:EditionID"`
	GameVersions     []gameVersionTable2V15 `gorm:"many2many:edition_game_version_relations;joinForeignKey:EditionID;joinReferences:GameVersionID"`
}

//nolint:unused
func (*editionTableV15) TableName() string {
	return "editions"
}

// V15
// v2_game_versionsの(GameID,Name)の組をuniqueに変更し，既存データの重複をリネームして反映する。
func v15() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "15",
		Migrate: func(tx *gorm.DB) error {
			var gameVersions []gameVersionTable2V15
			if err := tx.Order("game_id, name, created_at").
				Find(&gameVersions).Error; err != nil {
				return err
			}

			// 同一GameID内でnameの重複をリネーム
			var currentGameID uuid.UUID
			tmpMap := make(map[string]int)
			for _, gameVersion := range gameVersions {
				// 違うゲームになったらmapを初期化
				if currentGameID != gameVersion.GameID {
					currentGameID = gameVersion.GameID
					tmpMap = make(map[string]int)
				}

				if count, exists := tmpMap[gameVersion.Name]; exists {
					newName := gameVersion.Name + "+" + strconv.Itoa(count)
					tmpMap[gameVersion.Name] = count + 1
					gameVersion.Name = newName
				} else {
					tmpMap[gameVersion.Name] = 1
				}
				if err := tx.Save(&gameVersion).Error; err != nil {
					return err
				}
			}

			if err := tx.AutoMigrate(&gameVersionTable2V15{}); err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// 複合ユニーク制約を削除
			if err := tx.Exec(
				"ALTER TABLE v2_game_versions DROP INDEX unique_game_version_per_game",
			).Error; err != nil {
				return err
			}
			return nil
		},
	}
}
