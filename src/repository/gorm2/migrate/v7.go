package migrate

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// v7
// adminテーブルを追加
func v7() *gormigrate.Migration {
	tables := []any{
		&adminTable{},
	}
	return &gormigrate.Migration{
		ID: "7",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(tables...)
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(tables...)
		},
	}
}

type adminTable struct {
	UserID uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
}

//nolint:unused
func (*adminTable) TableName() string {
	return "admins"
}
