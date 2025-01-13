package migrate

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const (
	gameVideoTypeMkvV14 = "mkv"
	gameVideoTypeM4vV14 = "m4v"
)

func v14() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "14",
		Migrate: func(tx *gorm.DB) error {
			return tx.Create([]*gameVideoTypeTable{
				{
					Name:   gameVideoTypeMkvV14,
					Active: true,
				},
				{
					Name:   gameVideoTypeM4vV14,
					Active: true,
				},
			}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Where("name IN ?", []string{gameVideoTypeMkvV14, gameVideoTypeM4vV14}).Delete(&gameVideoTypeTable{}).Error
		},
	}
}
