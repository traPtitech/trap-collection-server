package gorm2

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSetupImageTypeTable(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		description      string
		beforeImageTypes []string
		isErr            bool
		err              error
	}

	testCases := []test{
		{
			description:      "何も存在しない場合問題なし",
			beforeImageTypes: []string{},
		},
		{
			description: "1つのみ存在する場合問題なし",
			beforeImageTypes: []string{
				gameImageTypeJpeg,
			},
		},
		{
			description: "2つ存在する場合問題なし",
			beforeImageTypes: []string{
				gameImageTypeJpeg,
				gameImageTypePng,
			},
		},
		{
			description: "全て存在する場合問題なし",
			beforeImageTypes: []string{
				gameImageTypeJpeg,
				gameImageTypePng,
				gameImageTypeGif,
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
					Delete(&GameImageTypeTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete role type table: %+v\n", err)
				}
			}()

			if len(testCase.beforeImageTypes) != 0 {
				imageTypes := make([]*GameImageTypeTable, 0, len(testCase.beforeImageTypes))
				for _, imageType := range testCase.beforeImageTypes {
					imageTypes = append(imageTypes, &GameImageTypeTable{
						Name: imageType,
					})
				}

				err := db.Create(imageTypes).Error
				if err != nil {
					t.Fatalf("failed to setup role type table: %+v\n", err)
				}
			}

			err := setupImageTypeTable(db)

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

			var imageTypes []*GameImageTypeTable
			err = db.
				Select("name").
				Find(&imageTypes).Error
			if err != nil {
				t.Fatalf("failed to get role type table: %+v\n", err)
			}

			imageTypeNames := make([]string, 0, len(imageTypes))
			for _, imageType := range imageTypes {
				imageTypeNames = append(imageTypeNames, imageType.Name)
			}

			assert.ElementsMatch(t, []string{
				gameImageTypeJpeg,
				gameImageTypePng,
				gameImageTypeGif,
			}, imageTypeNames)
		})
	}
}
