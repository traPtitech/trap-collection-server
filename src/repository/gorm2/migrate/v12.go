package migrate

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gameTable2V12 struct {
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
	// GameGenres
	// 後方参照を使っているためポインタになっている。
	// 参考: https://gorm.io/ja_JP/docs/many_to_many.html#%E5%BE%8C%E6%96%B9%E5%8F%82%E7%85%A7%EF%BC%88Back-Reference%EF%BC%89
	GameGenres []*gameGenreTableV12 `gorm:"many2many:game_genre_relations;joinForeignKey:GameID;joinReferences:GenreID"`
}

func (*gameTable2V12) TableName() string {
	return "games"
}

type gameGenreTableV12 struct {
	ID        uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	Name      string    `gorm:"type:varchar(32);not null;unique"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	// 後方参照を使っているためポインタになっている。
	// 参考: https://gorm.io/ja_JP/docs/many_to_many.html#%E5%BE%8C%E6%96%B9%E5%8F%82%E7%85%A7%EF%BC%88Back-Reference%EF%BC%89
	Games []*gameTable2V12 `gorm:"many2many:game_genre_relations;joinForeignKey:GenreID;joinReferences:GameID"`
}

func (*gameGenreTableV12) TableName() string {
	return "game_genres"
}

func v12() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "12",
		Migrate: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&gameTable2V12{})
			if err != nil {
				return err
			}
			err = tx.AutoMigrate(&gameGenreTableV12{})
			if err != nil {
				return err
			}
			return nil
		},
	}
}
