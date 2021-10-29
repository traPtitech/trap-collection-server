package gorm2

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSetupRoleTypeTable(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		description     string
		beforeRoleTypes []string
		isErr           bool
		err             error
	}

	testCases := []test{
		{
			description:     "何も存在しない場合問題なし",
			beforeRoleTypes: []string{},
		},
		{
			description: "administratorのみ存在する場合問題なし",
			beforeRoleTypes: []string{
				gameManagementRoleTypeAdministrator,
			},
		},
		{
			description: "collaboratorのみ存在する場合問題なし",
			beforeRoleTypes: []string{
				gameManagementRoleTypeCollaborator,
			},
		},
		{
			description: "administratorとcollaboratorが共に存在する場合問題なし",
			beforeRoleTypes: []string{
				gameManagementRoleTypeAdministrator,
				gameManagementRoleTypeCollaborator,
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
					Delete(&GameManagementRoleTypeTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete role type table: %+v\n", err)
				}
			}()

			if len(testCase.beforeRoleTypes) != 0 {
				roleTypes := make([]*GameManagementRoleTypeTable, 0, len(testCase.beforeRoleTypes))
				for _, roleType := range testCase.beforeRoleTypes {
					roleTypes = append(roleTypes, &GameManagementRoleTypeTable{
						Name: roleType,
					})
				}

				err := db.Create(roleTypes).Error
				if err != nil {
					t.Fatalf("failed to setup role type table: %+v\n", err)
				}
			}

			err := setupRoleTypeTable(db)

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

			var roleTypes []*GameManagementRoleTypeTable
			err = db.
				Select("name").
				Find(&roleTypes).Error
			if err != nil {
				t.Fatalf("failed to get role type table: %+v\n", err)
			}

			roleTypeNames := make([]string, 0, len(roleTypes))
			for _, roleType := range roleTypes {
				roleTypeNames = append(roleTypeNames, roleType.Name)
			}

			assert.ElementsMatch(t, []string{
				gameManagementRoleTypeAdministrator,
				gameManagementRoleTypeCollaborator,
			}, roleTypeNames)
		})
	}
}
