package gorm2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func TestAddGameManagementRoles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameManagementRoleRepository := NewGameManagementRole(testDB)

	type test struct {
		description string
		gameID      values.GameID
		userIDs     []values.TraPMemberID
		role        values.GameManagementRole
		expectRoles []migrate.GameManagementRoleTable
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

	var roleTypes []*migrate.GameManagementRoleTypeTable
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
			expectRoles: []migrate.GameManagementRoleTable{
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
			expectRoles: []migrate.GameManagementRoleTable{
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
			expectRoles: []migrate.GameManagementRoleTable{
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
			expectRoles: []migrate.GameManagementRoleTable{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&migrate.GameTable{
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

			var roles []migrate.GameManagementRoleTable
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

func TestUpdateGameManagementRole(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameManagementRoleRepository := NewGameManagementRole(testDB)

	type test struct {
		description string
		gameID      values.GameID
		userID      values.TraPMemberID
		role        values.GameManagementRole
		beforeGames []migrate.GameTable
		beforeRoles []migrate.GameManagementRoleTable
		expectRoles []migrate.GameManagementRoleTable
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()
	gameID7 := values.NewGameID()
	gameID8 := values.NewGameID()

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())
	userID3 := values.NewTrapMemberID(uuid.New())
	userID4 := values.NewTrapMemberID(uuid.New())
	userID5 := values.NewTrapMemberID(uuid.New())
	userID6 := values.NewTrapMemberID(uuid.New())
	userID7 := values.NewTrapMemberID(uuid.New())
	userID8 := values.NewTrapMemberID(uuid.New())

	var roleTypes []*migrate.GameManagementRoleTypeTable
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
			userID:      userID1,
			role:        values.GameManagementRoleAdministrator,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			beforeRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID1),
					UserID:     uuid.UUID(userID1),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
			expectRoles: []migrate.GameManagementRoleTable{
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
			userID:      userID2,
			role:        values.GameManagementRoleCollaborator,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			beforeRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID2),
					UserID:     uuid.UUID(userID2),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
				},
			},
			expectRoles: []migrate.GameManagementRoleTable{
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
			userID:      userID3,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			beforeRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID3),
					UserID:     uuid.UUID(userID3),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
			role:  100,
			isErr: true,
		},
		{
			description: "更新対象以外のユーザーのroleが存在しても問題なし",
			gameID:      gameID4,
			userID:      userID4,
			role:        values.GameManagementRoleAdministrator,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			beforeRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID4),
					UserID:     uuid.UUID(userID4),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
				{
					GameID:     uuid.UUID(gameID4),
					UserID:     uuid.UUID(userID5),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
			expectRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID4),
					UserID:     uuid.UUID(userID4),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
				},
				{
					GameID:     uuid.UUID(gameID4),
					UserID:     uuid.UUID(userID5),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
		},
		{
			description: "更新対象以外のゲームのroleが存在しても問題なし",
			gameID:      gameID5,
			userID:      userID6,
			role:        values.GameManagementRoleAdministrator,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
				{
					ID:          uuid.UUID(gameID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			beforeRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID5),
					UserID:     uuid.UUID(userID6),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
				{
					GameID:     uuid.UUID(gameID6),
					UserID:     uuid.UUID(userID6),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
			expectRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID5),
					UserID:     uuid.UUID(userID6),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
				},
				{
					GameID:     uuid.UUID(gameID6),
					UserID:     uuid.UUID(userID6),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
		},
		{
			description: "roleが事前に存在していないのでエラー",
			gameID:      gameID7,
			userID:      userID7,
			role:        values.GameManagementRoleAdministrator,
			beforeRoles: []migrate.GameManagementRoleTable{},
			expectRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID7),
					UserID:     uuid.UUID(userID7),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
				},
			},
			isErr: true,
			err:   repository.ErrNoRecordUpdated,
		},
		{
			// 実際には起きないが、念のため確認
			description: "既に更新後のroleが存在しているのでエラー",
			gameID:      gameID8,
			userID:      userID8,
			role:        values.GameManagementRoleAdministrator,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID8),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			beforeRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID8),
					UserID:     uuid.UUID(userID8),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
				},
			},
			expectRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID8),
					UserID:     uuid.UUID(userID8),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
				},
			},
			isErr: true,
			err:   repository.ErrNoRecordUpdated,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.beforeGames) != 0 {
				err := db.Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}
			}

			if len(testCase.beforeRoles) != 0 {
				err = db.Create(&testCase.beforeRoles).Error
				if err != nil {
					t.Fatalf("failed to create game management role table: %+v\n", err)
				}
			}

			err = gameManagementRoleRepository.UpdateGameManagementRole(ctx, testCase.gameID, testCase.userID, testCase.role)

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

			for _, expectRole := range testCase.expectRoles {
				var actualRole migrate.GameManagementRoleTable
				err = db.
					Where("game_id = ? and user_id = ?", expectRole.GameID, expectRole.UserID).
					First(&actualRole).Error
				if err != nil {
					t.Fatalf("failed to get game management role table: %+v\n", err)
				}

				assert.Equal(t, expectRole.RoleTypeID, actualRole.RoleTypeID)
			}
		})
	}
}

func TestRemoveGameManagementRole(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameManagementRoleRepository := NewGameManagementRole(testDB)

	type test struct {
		description string
		gameID      values.GameID
		userID      values.TraPMemberID
		beforeGames []migrate.GameTable
		beforeRoles []migrate.GameManagementRoleTable
		expectRoles []migrate.GameManagementRoleTable
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

	var roleTypes []*migrate.GameManagementRoleTypeTable
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
			userID:      userID1,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			beforeRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID1),
					UserID:     uuid.UUID(userID1),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
				},
			},
			expectRoles: []migrate.GameManagementRoleTable{},
		},
		{
			description: "削除対象以外のユーザーのroleが存在しても問題なし",
			gameID:      gameID2,
			userID:      userID2,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			beforeRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID2),
					UserID:     uuid.UUID(userID2),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
				{
					GameID:     uuid.UUID(gameID2),
					UserID:     uuid.UUID(userID3),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
			expectRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID2),
					UserID:     uuid.UUID(userID3),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
		},
		{
			description: "削除対象以外のゲームのroleが存在しても問題なし",
			gameID:      gameID3,
			userID:      userID4,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			beforeRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID3),
					UserID:     uuid.UUID(userID4),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
				{
					GameID:     uuid.UUID(gameID4),
					UserID:     uuid.UUID(userID4),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
			expectRoles: []migrate.GameManagementRoleTable{
				{
					GameID:     uuid.UUID(gameID4),
					UserID:     uuid.UUID(userID4),
					RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
				},
			},
		},
		{
			description: "roleが事前に存在していないのでエラー",
			gameID:      gameID5,
			userID:      userID5,
			beforeRoles: []migrate.GameManagementRoleTable{},
			expectRoles: []migrate.GameManagementRoleTable{},
			isErr:       true,
			err:         repository.ErrNoRecordDeleted,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.beforeGames) != 0 {
				err := db.Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}
			}

			if len(testCase.beforeRoles) != 0 {
				err = db.Create(&testCase.beforeRoles).Error
				if err != nil {
					t.Fatalf("failed to create game management role table: %+v\n", err)
				}
			}

			err = gameManagementRoleRepository.RemoveGameManagementRole(ctx, testCase.gameID, testCase.userID)

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

			for _, expectRole := range testCase.expectRoles {
				var actualRole migrate.GameManagementRoleTable
				err = db.
					Where("game_id = ? and user_id = ?", expectRole.GameID, expectRole.UserID).
					First(&actualRole).Error
				if err != nil {
					t.Fatalf("failed to get game management role table: %+v\n", err)
				}

				assert.Equal(t, expectRole.RoleTypeID, actualRole.RoleTypeID)
			}
		})
	}
}

func TestGetGameManagersByGameID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameManagementRoleRepository := NewGameManagementRole(testDB)

	type test struct {
		description        string
		gameID             values.GameID
		games              []migrate.GameTable
		expectUserAndRoles []*repository.UserIDAndManagementRole
		isErr              bool
		err                error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()
	gameID7 := values.NewGameID()

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())
	userID3 := values.NewTrapMemberID(uuid.New())
	userID4 := values.NewTrapMemberID(uuid.New())
	userID5 := values.NewTrapMemberID(uuid.New())
	userID6 := values.NewTrapMemberID(uuid.New())

	var roleTypes []*migrate.GameManagementRoleTypeTable
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
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID1),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			expectUserAndRoles: []*repository.UserIDAndManagementRole{
				{
					UserID: userID1,
					Role:   values.GameManagementRoleAdministrator,
				},
			},
		},
		{
			description: "roleがcollaboratorでも問題なし",
			gameID:      gameID7,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID7),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID6),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
						},
					},
				},
			},
			expectUserAndRoles: []*repository.UserIDAndManagementRole{
				{
					UserID: userID6,
					Role:   values.GameManagementRoleCollaborator,
				},
			},
		},
		{
			description: "roleが存在しなくても問題なし",
			gameID:      gameID2,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			expectUserAndRoles: []*repository.UserIDAndManagementRole{},
		},
		{
			// 実際にはあり得ないが念のため確認
			description:        "gameが存在しなくても問題なし",
			gameID:             gameID3,
			games:              []migrate.GameTable{},
			expectUserAndRoles: []*repository.UserIDAndManagementRole{},
		},
		{
			description: "roleが複数でも問題なし",
			gameID:      gameID4,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID2),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
						{
							UserID:     uuid.UUID(userID3),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
						},
					},
				},
			},
			expectUserAndRoles: []*repository.UserIDAndManagementRole{
				{
					UserID: userID2,
					Role:   values.GameManagementRoleAdministrator,
				},
				{
					UserID: userID3,
					Role:   values.GameManagementRoleCollaborator,
				},
			},
		},
		{
			description: "他のgameにroleがあっても問題なし",
			gameID:      gameID5,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID4),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
				{
					ID:          uuid.UUID(gameID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID5),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			expectUserAndRoles: []*repository.UserIDAndManagementRole{
				{
					UserID: userID4,
					Role:   values.GameManagementRoleAdministrator,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.games) != 0 {
				err = db.
					Session(&gorm.Session{}).
					Create(&testCase.games).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}
			}

			userAndRoles, err := gameManagementRoleRepository.GetGameManagersByGameID(ctx, testCase.gameID)

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

			assert.Len(t, userAndRoles, len(testCase.expectUserAndRoles))

			expectUserAndRoleMap := make(map[values.TraPMemberID]*repository.UserIDAndManagementRole, len(testCase.expectUserAndRoles))
			for _, userAndRole := range testCase.expectUserAndRoles {
				expectUserAndRoleMap[userAndRole.UserID] = userAndRole
			}

			for _, userAndRole := range userAndRoles {
				expectUserAndRole := expectUserAndRoleMap[userAndRole.UserID]
				assert.Equal(t, expectUserAndRole.UserID, userAndRole.UserID)
				assert.Equal(t, expectUserAndRole.Role, userAndRole.Role)
			}
		})
	}
}

func TestGetGameManagementRole(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameManagementRoleRepository := NewGameManagementRole(testDB)

	type test struct {
		description string
		gameID      values.GameID
		userID      values.TraPMemberID
		lockType    repository.LockType
		games       []migrate.GameTable
		role        values.GameManagementRole
		isErr       bool
		err         error
	}

	gameID1 := values.GameID(uuid.New())
	gameID2 := values.GameID(uuid.New())
	gameID3 := values.GameID(uuid.New())
	gameID4 := values.GameID(uuid.New())
	gameID5 := values.GameID(uuid.New())
	gameID6 := values.GameID(uuid.New())
	gameID7 := values.GameID(uuid.New())
	gameID8 := values.GameID(uuid.New())
	gameID9 := values.GameID(uuid.New())
	gameID10 := values.GameID(uuid.New())
	gameID11 := values.GameID(uuid.New())
	gameID12 := values.GameID(uuid.New())

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())
	userID3 := values.NewTrapMemberID(uuid.New())
	userID4 := values.NewTrapMemberID(uuid.New())
	userID5 := values.NewTrapMemberID(uuid.New())
	userID6 := values.NewTrapMemberID(uuid.New())
	userID7 := values.NewTrapMemberID(uuid.New())
	userID8 := values.NewTrapMemberID(uuid.New())
	userID9 := values.NewTrapMemberID(uuid.New())
	userID10 := values.NewTrapMemberID(uuid.New())
	userID11 := values.NewTrapMemberID(uuid.New())
	userID12 := values.NewTrapMemberID(uuid.New())

	var roleTypes []*migrate.GameManagementRoleTypeTable
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
			userID:      userID1,
			lockType:    repository.LockTypeNone,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID1),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			role: values.GameManagementRoleAdministrator,
		},
		{
			description: "roleがcollatorでも問題なし",
			gameID:      gameID2,
			userID:      userID2,
			lockType:    repository.LockTypeNone,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID2),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
						},
					},
				},
			},
			role: values.GameManagementRoleCollaborator,
		},
		{
			description: "roleが存在しないのでErrRecordNotFound",
			gameID:      gameID3,
			userID:      userID3,
			lockType:    repository.LockTypeNone,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			isErr: true,
			err:   repository.ErrRecordNotFound,
		},
		{
			// 実際にはgameIDのチェックが入り、行われることはないが念のため確認
			description: "gameIDが存在しないのでErrRecordNotFound",
			gameID:      gameID4,
			userID:      userID4,
			lockType:    repository.LockTypeNone,
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			description: "別のユーザーのroleがあっても問題なし",
			gameID:      gameID5,
			userID:      userID5,
			lockType:    repository.LockTypeNone,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID5),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
						{
							UserID:     uuid.UUID(userID6),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
						},
					},
				},
			},
			role: values.GameManagementRoleAdministrator,
		},
		{
			description: "別のユーザーのroleがあってもroleがなければErrRecordNotFound",
			gameID:      gameID6,
			userID:      userID7,
			lockType:    repository.LockTypeNone,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID8),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
						},
					},
				},
			},
			isErr: true,
			err:   repository.ErrRecordNotFound,
		},
		{
			description: "別のゲームのroleがあっても問題なし",
			gameID:      gameID7,
			userID:      userID9,
			lockType:    repository.LockTypeNone,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID7),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID9),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
				{
					ID:          uuid.UUID(gameID8),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID9),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
						},
					},
				},
			},
			role: values.GameManagementRoleAdministrator,
		},
		{
			description: "別のゲームのroleがあってもroleがなければErrRecordNotFound",
			gameID:      gameID9,
			userID:      userID10,
			lockType:    repository.LockTypeNone,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID9),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
				{
					ID:          uuid.UUID(gameID10),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID10),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
						},
					},
				},
			},
			isErr: true,
			err:   repository.ErrRecordNotFound,
		},
		{
			description: "行ロックでも問題なし",
			gameID:      gameID11,
			userID:      userID11,
			lockType:    repository.LockTypeRecord,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID11),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID11),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			role: values.GameManagementRoleAdministrator,
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ロックの種類が誤っているのでエラー",
			gameID:      gameID12,
			userID:      userID12,
			lockType:    100,
			games: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID12),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID12),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.games) != 0 {
				err = db.
					Session(&gorm.Session{}).
					Create(&testCase.games).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}
			}

			role, err := gameManagementRoleRepository.GetGameManagementRole(ctx, testCase.gameID, testCase.userID, testCase.lockType)

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

			assert.Equal(t, testCase.role, role)
		})
	}
}
