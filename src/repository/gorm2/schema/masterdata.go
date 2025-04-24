package schema

import (
	"context"
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func createMasterData(ctx context.Context, db *gorm.DB) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := createGameFileTypes(ctx, tx); err != nil {
			return fmt.Errorf("failed to create game file types: %w", err)
		}

		if err := createGameImageTypes(ctx, tx); err != nil {
			return fmt.Errorf("failed to create game image types: %w", err)
		}

		if err := createGameVideoTypes(ctx, tx); err != nil {
			return fmt.Errorf("failed to create game video types: %w", err)
		}

		if err := createGameManagementRoleTypes(ctx, tx); err != nil {
			return fmt.Errorf("failed to create game management role types: %w", err)
		}

		if err := createProductKeyStatusTypes(ctx, tx); err != nil {
			return fmt.Errorf("failed to create product key status types: %w", err)
		}

		if err := createSeatStatusTypes(ctx, tx); err != nil {
			return fmt.Errorf("failed to create seat status types: %w", err)
		}

		if err := createGameVisibilityTypes(ctx, tx); err != nil {
			return fmt.Errorf("failed to create game visibility type types: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create master data: %w", err)
	}

	return nil
}

func createGameFileTypes(_ context.Context, db *gorm.DB) error {
	gameFileTypes := []migrate.GameFileTypeTable{
		{Name: migrate.GameFileTypeJar, Active: true},
		{Name: migrate.GameFileTypeWindows, Active: true},
		{Name: migrate.GameFileTypeMac, Active: true},
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, gameFileType := range gameFileTypes {
			err := tx.Model(&migrate.GameFileTypeTable{}).FirstOrCreate(&migrate.GameFileTypeTable{}, gameFileType).Error
			if err != nil {
				return fmt.Errorf("failed to create game file type: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create game file types: %w", err)
	}

	return nil
}

func createGameImageTypes(_ context.Context, db *gorm.DB) error {
	gameImageTypes := []migrate.GameImageTypeTable{
		{Name: migrate.GameImageTypeJpeg, Active: true},
		{Name: migrate.GameImageTypePng, Active: true},
		{Name: migrate.GameImageTypeGif, Active: true},
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, gameImageType := range gameImageTypes {
			err := tx.Model(&migrate.GameImageTypeTable{}).FirstOrCreate(&migrate.GameImageTypeTable{}, gameImageType).Error
			if err != nil {
				return fmt.Errorf("failed to create game image type: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create game image types: %w", err)
	}

	return nil
}

func createGameVideoTypes(_ context.Context, db *gorm.DB) error {
	gameVideoTypes := []migrate.GameVideoTypeTable{
		{Name: migrate.GameVideoTypeMp4, Active: true},
		{Name: migrate.GameVideoTypeMkv, Active: true},
		{Name: migrate.GameVideoTypeM4v, Active: true},
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, gameVideoType := range gameVideoTypes {
			err := tx.FirstOrCreate(&migrate.GameVideoTypeTable{}, gameVideoType).Error
			if err != nil {
				return fmt.Errorf("failed to create game video type: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create game video types: %w", err)
	}

	return nil
}

func createGameManagementRoleTypes(_ context.Context, db *gorm.DB) error {
	gameManagementRoleTypes := []migrate.GameManagementRoleTypeTable{
		{Name: migrate.GameManagementRoleTypeAdministrator, Active: true},
		{Name: migrate.GameManagementRoleTypeCollaborator, Active: true},
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, gameManagementRoleType := range gameManagementRoleTypes {
			err := tx.FirstOrCreate(&migrate.GameManagementRoleTypeTable{}, gameManagementRoleType).Error
			if err != nil {
				return fmt.Errorf("failed to create game management role type: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create game management role types: %w", err)
	}

	return nil
}

func createProductKeyStatusTypes(_ context.Context, db *gorm.DB) error {
	productKeyStatusTypes := []migrate.ProductKeyStatusTable2{
		{Name: migrate.ProductKeyStatusActive, Active: true},
		{Name: migrate.ProductKeyStatusInactive, Active: true},
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, productKeyStatusType := range productKeyStatusTypes {
			err := tx.FirstOrCreate(&migrate.ProductKeyStatusTable2{}, productKeyStatusType).Error
			if err != nil {
				return fmt.Errorf("failed to create product key status type: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create product key status types: %w", err)
	}

	return nil
}

func createSeatStatusTypes(_ context.Context, db *gorm.DB) error {
	seatStatusTypes := []migrate.SeatStatusTable2{
		{Name: migrate.SeatStatusNone, Active: true},
		{Name: migrate.SeatStatusEmpty, Active: true},
		{Name: migrate.SeatStatusInUse, Active: true},
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, seatStatusType := range seatStatusTypes {
			err := tx.FirstOrCreate(&migrate.SeatStatusTable2{}, seatStatusType).Error
			if err != nil {
				return fmt.Errorf("failed to create seat status type: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create seat status types: %w", err)
	}

	return nil
}

func createGameVisibilityTypes(_ context.Context, db *gorm.DB) error {
	gameVisibilityTypes := []migrate.GameVisibilityTypeTable{
		{Name: migrate.GameVisibilityTypePublic},
		{Name: migrate.GameVisibilityTypeLimited},
		{Name: migrate.GameVisibilityTypePrivate},
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, gameVisibilityType := range gameVisibilityTypes {
			err := tx.FirstOrCreate(&migrate.GameVisibilityTypeTable{}, gameVisibilityType).Error
			if err != nil {
				return fmt.Errorf("failed to create game visibility type type: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create game visibility type types: %w", err)
	}

	return nil
}
