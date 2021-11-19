package gorm2

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSetupFileTypeTable(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		description     string
		beforeFileTypes []string
		isErr           bool
		err             error
	}

	testCases := []test{
		{
			description:     "何も存在しない場合問題なし",
			beforeFileTypes: []string{},
		},
		{
			description: "1つのみ存在する場合問題なし",
			beforeFileTypes: []string{
				gameFileTypeJar,
			},
		},
		{
			description: "2つ存在する場合問題なし",
			beforeFileTypes: []string{
				gameFileTypeJar,
				gameFileTypeWindows,
			},
		},
		{
			description: "全て存在する場合問題なし",
			beforeFileTypes: []string{
				gameFileTypeJar,
				gameFileTypeWindows,
				gameFileTypeMac,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Delete(&GameFileTypeTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete role type table: %+v\n", err)
				}
			}()

			if len(testCase.beforeFileTypes) != 0 {
				fileTypes := make([]*GameFileTypeTable, 0, len(testCase.beforeFileTypes))
				for _, fileType := range testCase.beforeFileTypes {
					fileTypes = append(fileTypes, &GameFileTypeTable{
						Name: fileType,
					})
				}

				err := db.Create(fileTypes).Error
				if err != nil {
					t.Fatalf("failed to setup role type table: %+v\n", err)
				}
			}

			err := setupFileTypeTable(db)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			var fileTypes []*GameFileTypeTable
			err = db.
				Select("name").
				Find(&fileTypes).Error
			if err != nil {
				t.Fatalf("failed to get role type table: %+v\n", err)
			}

			fileTypeNames := make([]string, 0, len(fileTypes))
			for _, fileType := range fileTypes {
				fileTypeNames = append(fileTypeNames, fileType.Name)
			}

			assert.ElementsMatch(t, []string{
				gameFileTypeJar,
				gameFileTypeWindows,
				gameFileTypeMac,
			}, fileTypeNames)
		})
	}
}
