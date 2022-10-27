package migrate

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// v9
// seatテーブル作成
func v9() *gormigrate.Migration {
	tables := []any{
		&seatTableV9{},
		&seatStatusTableV9{},
	}

	return &gormigrate.Migration{
		ID: "9",
		Migrate: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(tables...)
			if err != nil {
				return fmt.Errorf("failed to migrate: %w", err)
			}

			err = setupSeatStatusTableV9(tx)
			if err != nil {
				return fmt.Errorf("failed to setup seat status: %w", err)
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(tables...)
		},
	}
}

type seatTableV9 struct {
	ID         uint              `gorm:"type:int;primaryKey;not null"`
	StatusID   uint8             `gorm:"type:tinyint;not null"`
	SeatStatus seatStatusTableV9 `gorm:"foreignKey:StatusID"`
}

func (*seatTableV9) TableName() string {
	return "seats"
}

type seatStatusTableV9 struct {
	ID     uint8  `gorm:"type:tinyint;primaryKey;not null"`
	Name   string `gorm:"type:varchar(255);not null"`
	Active bool   `gorm:"type:boolean;not null;default:true"`
}

func (*seatStatusTableV9) TableName() string {
	return "seat_statuses"
}

const (
	seatStatusNoneV9  = "none"
	seatStatusEmptyV9 = "empty"
	seatStatusInUseV9 = "in_use"
)

func setupSeatStatusTableV9(db *gorm.DB) error {
	seatStatuses := []seatStatusTableV9{
		{
			Name:   seatStatusNoneV9,
			Active: true,
		},
		{
			Name:   seatStatusEmptyV9,
			Active: true,
		},
		{
			Name:   seatStatusInUseV9,
			Active: true,
		},
	}

	for _, status := range seatStatuses {
		err := db.
			Session(&gorm.Session{}).
			Where("name = ?", status.Name).
			FirstOrCreate(&status).Error
		if err != nil {
			return fmt.Errorf("failed to create seat status: %w", err)
		}
	}

	return nil
}
