package migrate

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// v6
// product_keyテーブルからdeleted_atを削除し、statusを追加
func v6() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "6",
		Migrate: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&productKeyStatusTableV6{})
			if err != nil {
				return fmt.Errorf("failed to migrate product_key_status table: %w", err)
			}

			err = setupProductKeyStatusTableV6(tx)
			if err != nil {
				return fmt.Errorf("failed to setup product key status table: %w", err)
			}

			var status productKeyStatusTableV6
			err = tx.
				Session(&gorm.Session{}).
				Where("name = ?", productKeyStatusActiveV6).
				Take(&status).Error
			if err != nil {
				return fmt.Errorf("failed to get product key status: %w", err)
			}

			err = tx.Exec(fmt.Sprintf("ALTER TABLE product_keys ADD COLUMN status_id tinyint(4) NOT NULL DEFAULT %d", status.ID)).Error
			if err != nil {
				return fmt.Errorf("failed to add status_id column: %w", err)
			}

			err = tx.Exec("UPDATE product_keys SET status_id = (SELECT id FROM product_key_statuses WHERE name = ?) WHERE deleted_at IS NOT NULL", productKeyStatusInactiveV6).Error
			if err != nil {
				return fmt.Errorf("failed to update status_id column: %w", err)
			}

			err = tx.AutoMigrate(&productKeyTableV6{})
			if err != nil {
				return fmt.Errorf("failed to migrate product_key table: %w", err)
			}

			err = tx.Migrator().DropColumn(&productKeyTableV2{}, "deleted_at")
			if err != nil {
				return fmt.Errorf("failed to drop deleted_at column: %w", err)
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&productKeyTableV2{})
			if err != nil {
				return fmt.Errorf("failed to migrate product_key table: %w", err)
			}

			err = tx.Exec("UPDATE product_keys JOIN product_key_statuses ON product_keys.status_id = product_key_statuses.id SET product_keys.deleted_at =  SET deleted_at = CURRENT_TIMESTAMP WHERE product_key_statuses.name = ?", productKeyStatusInactiveV6).Error
			if err != nil {
				return fmt.Errorf("failed to set deleted_at: %w", err)
			}

			err = tx.Migrator().DropColumn(&productKeyTableV6{}, "status_id")
			if err != nil {
				return fmt.Errorf("failed to drop status_id column: %w", err)
			}

			err = tx.Migrator().DropTable(&productKeyStatusTableV6{})
			if err != nil {
				return fmt.Errorf("failed to drop product_key_status_types table: %w", err)
			}

			return nil
		},
	}
}

type productKeyTableV6 struct {
	ID           uuid.UUID               `gorm:"type:varchar(36);not null;primaryKey"`
	EditionID    uuid.UUID               `gorm:"type:varchar(36);not null"`
	ProductKey   string                  `gorm:"type:varchar(29);not null;unique"`
	StatusID     int                     `gorm:"type:tinyint;not null"`
	CreatedAt    time.Time               `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	Status       productKeyStatusTableV6 `gorm:"foreignKey:StatusID"`
	AccessTokens []accessTokenTableV2    `gorm:"foreignKey:ProductKeyID"`
}

func (*productKeyTableV6) TableName() string {
	return "product_keys"
}

type productKeyStatusTableV6 struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*productKeyStatusTableV6) TableName() string {
	return "product_key_statuses"
}

const (
	productKeyStatusActiveV6   = "active"   // 有効
	productKeyStatusInactiveV6 = "inactive" /// 無効
)

func setupProductKeyStatusTableV6(db *gorm.DB) error {
	statusList := []productKeyStatusTableV6{
		{
			Name:   productKeyStatusActiveV6,
			Active: true,
		},
		{
			Name:   productKeyStatusInactiveV6,
			Active: true,
		},
	}

	for _, status := range statusList {
		err := db.
			Session(&gorm.Session{}).
			Where("name = ?", status.Name).
			FirstOrCreate(&status).Error
		if err != nil {
			return fmt.Errorf("failed to create product key status type: %w", err)
		}
	}

	return nil
}

type editionTableV6 struct {
	ID               uuid.UUID             `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string                `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString        `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time             `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt        `gorm:"type:DATETIME NULL;default:NULL"`
	ProductKeys      []productKeyTableV6   `gorm:"foreignKey:EditionID"`
	GameVersions     []gameVersionTable2V5 `gorm:"many2many:edition_game_version_relations;joinForeignKey:EditionID;joinReferences:GameVersionID"`
}

func (*editionTableV6) TableName() string {
	return "editions"
}
