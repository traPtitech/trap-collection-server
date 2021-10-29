package gorm2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
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

func TestAddGameManagementRoles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameManagementRoleRepository, err := NewGameManagementRole(testDB)
	if err != nil {
		t.Fatalf("failed to create game management role repository: %+v\n", err)
	}

	type test struct {
		description string
		gameID      values.GameID
		userIDs     []values.TraPMemberID
		role        values.GameManagementRole
		expectRoles []GameManagementRoleTable
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())
	userID3 := values.NewTrapMemberID(uuid.New())
	userID4 := values.NewTrapMemberID(uuid.New())
	userID5 := values.NewTrapMemberID(uuid.New())

	var roleTypes []*GameManagementRoleTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&roleTypes).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	roleTypeMap := make(map[string]int, len(roleTypes))
	for _, roleType := range roleTypes {
		roleTypeMap[roleType.Name] = roleType.ID
	}

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			gameID:      gameID1,
			userIDs: []values.TraPMemberID{
				userID1,
			},
			role: values.GameManagementRoleAdministrator,
			expectRoles: []GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID1),
					UserID:     uuid.UUID(userID1),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
				},
			},
		},
		{
			description: "roleがCollaboratorでも問題なし",
			gameID:      gameID2,
			userIDs: []values.TraPMemberID{
				userID2,
			},
			role: values.GameManagementRoleCollaborator,
			expectRoles: []GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID2),
					UserID:     uuid.UUID(userID2),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
		},
		{
			description: "roleがAdministratorでもCollaboratorでもないのでエラー",
			gameID:      gameID3,
			userIDs: []values.TraPMemberID{
				userID3,
			},
			role:  100,
			isErr: true,
		},
		{
			description: "ユーザーが複数でも問題なし",
			gameID:      gameID4,
			userIDs: []values.TraPMemberID{
				userID4,
				userID5,
			},
			role: values.GameManagementRoleAdministrator,
			expectRoles: []GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID4),
					UserID:     uuid.UUID(userID4),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
				},
				{
					GameID:     uuid.UUID(gameID4),
					UserID:     uuid.UUID(userID5),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
				},
			},
		},
		{
			description: "ユーザーが0人でも問題なし",
			gameID:      gameID5,
			userIDs:     []values.TraPMemberID{},
			role:        values.GameManagementRoleAdministrator,
			expectRoles: []GameManagementRoleTable{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&GameTable{
				ID:          uuid.UUID(testCase.gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   time.Now(),
			}).Error
			if err != nil {
				t.Fatalf("failed to create game table: %+v\n", err)
			}

			err = gameManagementRoleRepository.AddGameManagementRoles(ctx, testCase.gameID, testCase.userIDs, testCase.role)

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

			var roles []GameManagementRoleTable
			err = db.
				Session(&gorm.Session{}).
				Where("game_id = ?", uuid.UUID(testCase.gameID)).
				Find(&roles).Error
			if err != nil {
				t.Fatalf("failed to get role table: %+v\n", err)
			}

			assert.ElementsMatch(t, testCase.expectRoles, roles)
		})
	}
}
