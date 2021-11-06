package gorm2

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSetupVideoTypeTable(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		description      string
		beforeVideoTypes []string
		isErr            bool
		err              error
	}

	testCases := []test{
		{
			description:      "何も存在しない場合問題なし",
			beforeVideoTypes: []string{},
		},
		{
			description: "全て存在する場合問題なし",
			beforeVideoTypes: []string{
				gameVideoTypeMp4,
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
					Delete(&GameVideoTypeTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete role type table: %+v\n", err)
				}
			}()

			if len(testCase.beforeVideoTypes) != 0 {
				videoTypes := make([]*GameVideoTypeTable, 0, len(testCase.beforeVideoTypes))
				for _, videoType := range testCase.beforeVideoTypes {
					videoTypes = append(videoTypes, &GameVideoTypeTable{
						Name: videoType,
					})
				}

				err := db.Create(videoTypes).Error
				if err != nil {
					t.Fatalf("failed to setup role type table: %+v\n", err)
				}
			}

			err := setupVideoTypeTable(db)

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

			var videoTypes []*GameVideoTypeTable
			err = db.
				Select("name").
				Find(&videoTypes).Error
			if err != nil {
				t.Fatalf("failed to get role type table: %+v\n", err)
			}

			videoTypeNames := make([]string, 0, len(videoTypes))
			for _, videoType := range videoTypes {
				videoTypeNames = append(videoTypeNames, videoType.Name)
			}

			assert.ElementsMatch(t, []string{
				gameVideoTypeMp4,
			}, videoTypeNames)
		})
	}
}
