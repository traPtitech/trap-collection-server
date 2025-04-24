package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func Test_createGameFileTypes(t *testing.T) {
	db := testDB.WithContext(t.Context())
	expectedGameFileTypes := map[string]migrate.GameFileTypeTable{
		migrate.GameFileTypeJar:     {Name: migrate.GameFileTypeJar, Active: true},
		migrate.GameFileTypeWindows: {Name: migrate.GameFileTypeWindows, Active: true},
		migrate.GameFileTypeMac:     {Name: migrate.GameFileTypeMac, Active: true},
	}

	var initialGameFileTypes []migrate.GameFileTypeTable
	err := db.Find(&initialGameFileTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, len(expectedGameFileTypes), len(initialGameFileTypes))
	for _, data := range initialGameFileTypes {
		want, ok := expectedGameFileTypes[data.Name]
		assert.True(t, ok)
		assert.Equal(t, want.Name, data.Name)
	}

	err = createGameFileTypes(t.Context(), db)
	assert.NoError(t, err)

	var newGameFileTypes []migrate.GameFileTypeTable
	err = db.Find(&newGameFileTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, initialGameFileTypes, newGameFileTypes)
}

func Test_createGameImageTypes(t *testing.T) {
	db := testDB.WithContext(t.Context())
	expectedGameImageTypes := map[string]migrate.GameImageTypeTable{
		migrate.GameImageTypeJpeg: {Name: migrate.GameImageTypeJpeg, Active: true},
		migrate.GameImageTypePng:  {Name: migrate.GameImageTypePng, Active: true},
		migrate.GameImageTypeGif:  {Name: migrate.GameImageTypeGif, Active: true},
	}

	var initialGameImageTypes []migrate.GameImageTypeTable
	err := db.Session(&gorm.Session{}).Find(&initialGameImageTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, len(expectedGameImageTypes), len(initialGameImageTypes))
	for _, data := range initialGameImageTypes {
		want, ok := expectedGameImageTypes[data.Name]
		assert.True(t, ok)
		assert.Equal(t, want.Name, data.Name)
		assert.Equal(t, want.Active, data.Active)
	}

	err = createGameImageTypes(t.Context(), db)
	assert.NoError(t, err)

	var newGameImageTypes []migrate.GameImageTypeTable
	err = db.Find(&newGameImageTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, initialGameImageTypes, newGameImageTypes)
}

func Test_createGameVideoTypes(t *testing.T) {
	db := testDB.WithContext(t.Context())

	expectedGameVideoTypes := map[string]migrate.GameVideoTypeTable{
		migrate.GameVideoTypeMp4: {Name: migrate.GameVideoTypeMp4, Active: true},
		migrate.GameVideoTypeM4v: {Name: migrate.GameVideoTypeM4v, Active: true},
		migrate.GameVideoTypeMkv: {Name: migrate.GameVideoTypeMkv, Active: true},
	}

	var initialGameVideoTypes []migrate.GameVideoTypeTable
	err := db.Find(&initialGameVideoTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, len(expectedGameVideoTypes), len(initialGameVideoTypes))
	for _, data := range initialGameVideoTypes {
		want, ok := expectedGameVideoTypes[data.Name]
		assert.True(t, ok)
		assert.Equal(t, want.Name, data.Name)
		assert.Equal(t, want.Active, data.Active)
	}

	err = createGameVideoTypes(t.Context(), db)
	assert.NoError(t, err)

	var newGameVideoTypes []migrate.GameVideoTypeTable
	err = db.Find(&newGameVideoTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, initialGameVideoTypes, newGameVideoTypes)
}

func Test_createGameManagementRoleTypes(t *testing.T) {
	db := testDB.WithContext(t.Context())

	expectedGameManagementRoleTypes := map[string]migrate.GameManagementRoleTypeTable{
		migrate.GameManagementRoleTypeAdministrator: {Name: migrate.GameManagementRoleTypeAdministrator, Active: true},
		migrate.GameManagementRoleTypeCollaborator:  {Name: migrate.GameManagementRoleTypeCollaborator, Active: true},
	}

	var initialGameManagementRoleTypes []migrate.GameManagementRoleTypeTable
	err := db.Find(&initialGameManagementRoleTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, len(expectedGameManagementRoleTypes), len(initialGameManagementRoleTypes))
	for _, data := range initialGameManagementRoleTypes {
		want, ok := expectedGameManagementRoleTypes[data.Name]
		assert.True(t, ok)
		assert.Equal(t, want.Name, data.Name)
		assert.Equal(t, want.Active, data.Active)
	}

	err = createGameManagementRoleTypes(t.Context(), db)
	assert.NoError(t, err)

	var newGameManagementRoleTypes []migrate.GameManagementRoleTypeTable
	err = db.Find(&newGameManagementRoleTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, initialGameManagementRoleTypes, newGameManagementRoleTypes)
}

func Test_createProductKeyStatusTypes(t *testing.T) {
	db := testDB.WithContext(t.Context())

	expectedProductKeyStatusTypes := map[string]migrate.ProductKeyStatusTable2{
		migrate.ProductKeyStatusActive:   {Name: migrate.ProductKeyStatusActive, Active: true},
		migrate.ProductKeyStatusInactive: {Name: migrate.ProductKeyStatusInactive, Active: true},
	}

	var initialProductKeyStatusTypes []migrate.ProductKeyStatusTable2
	err := db.Find(&initialProductKeyStatusTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, len(expectedProductKeyStatusTypes), len(initialProductKeyStatusTypes))
	for _, data := range initialProductKeyStatusTypes {
		want, ok := expectedProductKeyStatusTypes[data.Name]
		assert.True(t, ok)
		assert.Equal(t, want.Name, data.Name)
		assert.Equal(t, want.Active, data.Active)
	}

	err = createProductKeyStatusTypes(t.Context(), db)
	assert.NoError(t, err)

	var newProductKeyStatusTypes []migrate.ProductKeyStatusTable2
	err = db.Find(&newProductKeyStatusTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, initialProductKeyStatusTypes, newProductKeyStatusTypes)
}

func Test_createSeatStatusTypes(t *testing.T) {
	db := testDB.WithContext(t.Context())

	expectedSeatStatusTypes := map[string]migrate.SeatStatusTable2{
		migrate.SeatStatusEmpty: {Name: migrate.SeatStatusEmpty, Active: true},
		migrate.SeatStatusInUse: {Name: migrate.SeatStatusInUse, Active: true},
		migrate.SeatStatusNone:  {Name: migrate.SeatStatusNone, Active: true},
	}

	var initialSeatStatusTypes []migrate.SeatStatusTable2
	err := db.Find(&initialSeatStatusTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, len(expectedSeatStatusTypes), len(initialSeatStatusTypes))
	for _, data := range initialSeatStatusTypes {
		want, ok := expectedSeatStatusTypes[data.Name]
		assert.True(t, ok)
		assert.Equal(t, want.Name, data.Name)
		assert.Equal(t, want.Active, data.Active)
	}

	err = createSeatStatusTypes(t.Context(), db)
	assert.NoError(t, err)

	var newSeatStatusTypes []migrate.SeatStatusTable2
	err = db.Find(&newSeatStatusTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, initialSeatStatusTypes, newSeatStatusTypes)
}

func Test_createGameVisibilityTypes(t *testing.T) {
	db := testDB.WithContext(t.Context())

	expectedGameVisibilityTypes := map[string]migrate.GameVisibilityTypeTable{
		migrate.GameVisibilityTypePublic:  {Name: migrate.GameVisibilityTypePublic},
		migrate.GameVisibilityTypePrivate: {Name: migrate.GameVisibilityTypePrivate},
		migrate.GameVisibilityTypeLimited: {Name: migrate.GameVisibilityTypeLimited},
	}

	var initialGameVisibilityTypes []migrate.GameVisibilityTypeTable
	err := db.Find(&initialGameVisibilityTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, len(expectedGameVisibilityTypes), len(initialGameVisibilityTypes))
	for _, data := range initialGameVisibilityTypes {
		want, ok := expectedGameVisibilityTypes[data.Name]
		assert.True(t, ok)
		assert.Equal(t, want.Name, data.Name)
	}

	err = createGameVisibilityTypes(t.Context(), db)
	assert.NoError(t, err)

	var newGameVisibilityTypes []migrate.GameVisibilityTypeTable
	err = db.Find(&newGameVisibilityTypes).Error
	assert.NoError(t, err)
	assert.Equal(t, initialGameVisibilityTypes, newGameVisibilityTypes)
}
