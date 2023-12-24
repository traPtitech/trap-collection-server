package migrate

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gameGenreTableV10 struct {
	ID        uuid.UUID      `gorm:"type:varchar(36);not null;primaryKey"`
	Name      string         `gorm:"type:varchar(32);not null;unique"`
	CreatedAt time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	Games     []gameTable2V5 `gorm:"many2many:game_genre_relations;joinForeignKey:GenreID;joinReferences:GameID"`
}

//nolint:unused
func (*gameGenreTableV10) TableName() string {
	return "game_genres"
}

func v10() *gormigrate.Migration {
	tables := []any{
		&gameGenreTableV10{},
	}

	return &gormigrate.Migration{
		ID: "10",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(tables...)
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(tables...)
		},
	}
}
