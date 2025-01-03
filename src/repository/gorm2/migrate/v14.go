package migrate

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gameVersionTable2V14 struct {
	ID          uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	GameID      uuid.UUID `gorm:"type:varchar(36);not null;index:idx_game_id_name,unique"` // GameIDとNameの組み合わせでuniqueに
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

func (*gameVersionTable2V14) TableName() string {
	return "v2_game_versions"
}

//nolint:unused
type gameTable2V14 struct {
	ID                  uuid.UUID                 `gorm:"type:varchar(36);not null;primaryKey"`
	Name                string                    `gorm:"type:varchar(256);size:256;not null"`
	Description         string                    `gorm:"type:text;not null"`
	CreatedAt           time.Time                 `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt            `gorm:"type:DATETIME NULL;default:NULL"`
	GameVersions        []gameVersionTable2V14    `gorm:"foreignkey:GameID"`
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
func (*gameTable2V14) TableName() string {
	return "games"
}

//nolint:unused
type editionTableV14 struct {
	ID               uuid.UUID              `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string                 `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString         `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time              `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt         `gorm:"type:DATETIME NULL;default:NULL"`
	ProductKeys      []productKeyTableV2    `gorm:"foreignKey:EditionID"`
	GameVersions     []gameVersionTable2V14 `gorm:"many2many:edition_game_version_relations;joinForeignKey:EditionID;joinReferences:GameVersionID"`
}

//nolint:unused
func (*editionTableV14) TableName() string {
	return "editions"
}

// V14
// v2_game_versionsの(GameID,Name)の組をuniqueに変更し，既存データの重複をリネームして反映する。
func v14() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "14",
		Migrate: func(tx *gorm.DB) error {
			var gameVersions []gameVersionTable2V14
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

			// 複合ユニーク制約を追加
			if err := tx.Exec(
				"ALTER TABLE v2_game_versions ADD CONSTRAINT unique_game_version_per_game UNIQUE (game_id, name)",
			).Error; err != nil {
				return err
			}

			if err := tx.AutoMigrate(&gameVersionTable2V14{}); err != nil {
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
			// テーブル定義をロールバック
			if err := tx.Migrator().DropTable(&gameVersionTable2V14{}); err != nil {
				return err
			}
			return nil
		},
	}
}
