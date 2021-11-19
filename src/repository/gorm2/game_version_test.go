package gorm2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"gorm.io/gorm"
)

func TestCreateGameVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %v", err)
	}

	gameVersionRepository := NewGameVersion(testDB)

	type test struct {
		description        string
		isGameExist        bool
		isGameDeleted      bool
		gameID             values.GameID
		version            *domain.GameVersion
		beforeGameVersions []GameVersionTable
		expectGameVersions []GameVersionTable
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
	gameID8 := values.NewGameID()
	gameID9 := values.NewGameID()
	gameID10 := values.NewGameID()

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()
	gameVersionID7 := values.NewGameVersionID()
	gameVersionID8 := values.NewGameVersionID()
	gameVersionID9 := values.NewGameVersionID()
	gameVersionID10 := values.NewGameVersionID()
	gameVersionID11 := values.NewGameVersionID()
	gameVersionID12 := values.NewGameVersionID()

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			isGameExist: true,
			gameID:      gameID1,
			version: domain.NewGameVersion(
				gameVersionID1,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			beforeGameVersions: []GameVersionTable{},
			expectGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID1),
					GameID:      uuid.UUID(gameID1),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "既にバージョンが存在してもエラーなし",
			isGameExist: true,
			gameID:      gameID2,
			version: domain.NewGameVersion(
				gameVersionID2,
				values.NewGameVersionName("v1.1.0"),
				values.NewGameVersionDescription("アップデート"),
				now,
			),
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID3),
					GameID:      uuid.UUID(gameID2),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			expectGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID2),
					GameID:      uuid.UUID(gameID2),
					Name:        "v1.1.0",
					Description: "アップデート",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameVersionID3),
					GameID:      uuid.UUID(gameID2),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
		},
		{
			// 実際には発生しないが、念のため確認
			description: "既にIDが同じバージョンが存在するのでエラー",
			isGameExist: true,
			gameID:      gameID3,
			version: domain.NewGameVersion(
				gameVersionID4,
				values.NewGameVersionName("v1.1.0"),
				values.NewGameVersionDescription("アップデート"),
				now,
			),
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID4),
					GameID:      uuid.UUID(gameID3),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			expectGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID4),
					GameID:      uuid.UUID(gameID3),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			isErr: true,
		},
		{
			description: "同名のバージョンが存在してもエラーなし",
			isGameExist: true,
			gameID:      gameID4,
			version: domain.NewGameVersion(
				gameVersionID5,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("アップデート"),
				now,
			),
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID6),
					GameID:      uuid.UUID(gameID4),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			expectGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID5),
					GameID:      uuid.UUID(gameID4),
					Name:        "v1.0.0",
					Description: "アップデート",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameVersionID6),
					GameID:      uuid.UUID(gameID4),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
		},
		{
			description: "バージョン名が32文字でもエラーなし",
			isGameExist: true,
			gameID:      gameID5,
			version: domain.NewGameVersion(
				gameVersionID7,
				values.NewGameVersionName("v1.0.123456789012345678901234567"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			beforeGameVersions: []GameVersionTable{},
			expectGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID7),
					GameID:      uuid.UUID(gameID5),
					Name:        "v1.0.123456789012345678901234567",
					Description: "リリース",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "バージョン名が33文字なのでエラー",
			isGameExist: true,
			gameID:      gameID6,
			version: domain.NewGameVersion(
				gameVersionID8,
				values.NewGameVersionName("v1.0.1234567890123456789012345678"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			beforeGameVersions: []GameVersionTable{},
			expectGameVersions: []GameVersionTable{},
			isErr:              true,
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ゲームが存在しないのでエラー",
			gameID:      gameID7,
			version: domain.NewGameVersion(
				gameVersionID9,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			beforeGameVersions: []GameVersionTable{},
			expectGameVersions: []GameVersionTable{},
			isErr:              true,
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ゲームが削除されているのでエラー",
			gameID:      gameID8,
			version: domain.NewGameVersion(
				gameVersionID10,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			beforeGameVersions: []GameVersionTable{},
			expectGameVersions: []GameVersionTable{},
			isErr:              true,
		},
		{
			// 実際には発生しないが、念のため確認
			description: "バージョン名が空文字でもエラーなし",
			isGameExist: true,
			gameID:      gameID9,
			version: domain.NewGameVersion(
				gameVersionID11,
				values.NewGameVersionName(""),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			beforeGameVersions: []GameVersionTable{},
			expectGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID11),
					GameID:      uuid.UUID(gameID9),
					Name:        "",
					Description: "リリース",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "説明が空文字でもエラーなし",
			isGameExist: true,
			gameID:      gameID10,
			version: domain.NewGameVersion(
				gameVersionID12,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription(""),
				now,
			),
			beforeGameVersions: []GameVersionTable{},
			expectGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID12),
					GameID:      uuid.UUID(gameID10),
					Name:        "v1.0.0",
					Description: "",
					CreatedAt:   now,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isGameExist {
				err := db.
					Session(&gorm.Session{}).
					Create(&GameTable{
						ID:           uuid.UUID(testCase.gameID),
						Name:         "test",
						Description:  "test",
						CreatedAt:    time.Now(),
						GameVersions: testCase.beforeGameVersions,
					}).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}

				if testCase.isGameDeleted {
					err := db.
						Session(&gorm.Session{}).
						Where("id = ?", uuid.UUID(testCase.gameID)).
						Delete(&GameTable{}).Error
					if err != nil {
						t.Fatalf("failed to delete game table: %+v\n", err)
					}
				}
			}

			err := gameVersionRepository.CreateGameVersion(ctx, testCase.gameID, testCase.version)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			var gameVersions []GameVersionTable
			err = db.
				Where("game_id = ?", uuid.UUID(testCase.gameID)).
				Find(&gameVersions).Error
			if err != nil {
				t.Fatalf("failed to get game versions: %+v\n", err)
			}

			assert.Len(t, gameVersions, len(testCase.expectGameVersions))

			versionMap := make(map[uuid.UUID]GameVersionTable, len(gameVersions))
			for _, version := range gameVersions {
				versionMap[version.ID] = version
			}

			for _, expectVersion := range testCase.expectGameVersions {
				actualVersion, ok := versionMap[expectVersion.ID]
				if !ok {
					t.Errorf("failed to find version: %s", expectVersion.Name)
					continue
				}

				assert.Equal(t, expectVersion.ID, actualVersion.ID)
				assert.Equal(t, expectVersion.GameID, actualVersion.GameID)
				assert.Equal(t, expectVersion.Name, actualVersion.Name)
				assert.Equal(t, expectVersion.Description, actualVersion.Description)
				assert.WithinDuration(t, expectVersion.CreatedAt, actualVersion.CreatedAt, 2*time.Second)
			}
		})
	}
}

func TestGetGameVersions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %v", err)
	}

	gameVersionRepository := NewGameVersion(testDB)

	type test struct {
		description string
		gameID      values.GameID
		games       []GameTable
		expect      []*domain.GameVersion
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

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID1),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
						},
					},
				},
			},
			expect: []*domain.GameVersion{
				domain.NewGameVersion(
					gameVersionID1,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					time.Now(),
				),
			},
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ゲームが存在しなくてもエラーなし",
			gameID:      gameID2,
			games:       []GameTable{},
			expect:      []*domain.GameVersion{},
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ゲームが削除されていてもエラーなし",
			gameID:      gameID3,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					DeletedAt: gorm.DeletedAt{
						Time:  time.Now(),
						Valid: true,
					},
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID2),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
						},
					},
				},
			},
			expect: []*domain.GameVersion{},
		},
		{
			description: "バージョンが複数あってもエラーなし",
			gameID:      gameID4,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID3),
							Name:        "v1.1.0",
							Description: "アップデート",
							CreatedAt:   time.Now(),
						},
						{
							ID:          uuid.UUID(gameVersionID4),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now().Add(-time.Hour),
						},
					},
				},
			},
			expect: []*domain.GameVersion{
				domain.NewGameVersion(
					gameVersionID3,
					values.NewGameVersionName("v1.1.0"),
					values.NewGameVersionDescription("アップデート"),
					time.Now(),
				),
				domain.NewGameVersion(
					gameVersionID4,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					time.Now().Add(-time.Hour),
				),
			},
		},
		{
			description: "バージョンが存在しなくてもエラーなし",
			gameID:      gameID5,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			expect: []*domain.GameVersion{},
		},
		{
			description: "別のゲームのバージョンが混ざることはない",
			gameID:      gameID6,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID5),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
						},
					},
				},
				{
					ID:          uuid.UUID(gameID7),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID6),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
						},
					},
				},
			},
			expect: []*domain.GameVersion{
				domain.NewGameVersion(
					gameVersionID5,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					time.Now(),
				),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.games) != 0 {
				err := db.Create(&testCase.games).Error
				if err != nil {
					t.Fatalf("failed to create games: %v", err)
				}

				for _, game := range testCase.games {
					if game.DeletedAt.Valid {
						err := db.Delete(&game).Error
						if err != nil {
							t.Fatalf("failed to delete game: %v", err)
						}
					}
				}
			}

			gameVersions, err := gameVersionRepository.GetGameVersions(ctx, testCase.gameID)

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

			for i, expectVersion := range testCase.expect {
				actualVersion := gameVersions[i]

				assert.Equal(t, expectVersion.GetID(), actualVersion.GetID())
				assert.Equal(t, expectVersion.GetName(), actualVersion.GetName())
				assert.Equal(t, expectVersion.GetDescription(), actualVersion.GetDescription())
				assert.WithinDuration(t, expectVersion.GetCreatedAt(), actualVersion.GetCreatedAt(), 2*time.Second)
			}
		})
	}
}

func TestGetLatestGameVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %v", err)
	}

	gameVersionRepository := NewGameVersion(testDB)

	type test struct {
		description string
		gameID      values.GameID
		lockType    repository.LockType
		games       []GameTable
		expect      *domain.GameVersion
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

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()
	gameVersionID7 := values.NewGameVersionID()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			lockType:    repository.LockTypeNone,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID1),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
						},
					},
				},
			},
			expect: domain.NewGameVersion(
				gameVersionID1,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				time.Now(),
			),
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ゲームが存在しないのでErrRecordNotFound",
			gameID:      gameID2,
			lockType:    repository.LockTypeNone,
			games:       []GameTable{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ゲームが削除されていてもエラーなし",
			gameID:      gameID3,
			lockType:    repository.LockTypeNone,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					DeletedAt: gorm.DeletedAt{
						Time:  time.Now(),
						Valid: true,
					},
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID2),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
						},
					},
				},
			},
			expect: domain.NewGameVersion(
				gameVersionID2,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				time.Now(),
			),
		},
		{
			description: "バージョンが複数あってもエラーなし",
			gameID:      gameID4,
			lockType:    repository.LockTypeNone,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID3),
							Name:        "v1.1.0",
							Description: "アップデート",
							CreatedAt:   time.Now(),
						},
						{
							ID:          uuid.UUID(gameVersionID4),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now().Add(-time.Hour),
						},
					},
				},
			},
			expect: domain.NewGameVersion(
				gameVersionID3,
				values.NewGameVersionName("v1.1.0"),
				values.NewGameVersionDescription("アップデート"),
				time.Now(),
			),
		},
		{
			description: "バージョンが存在しないのでErrRecordNotFound",
			gameID:      gameID5,
			lockType:    repository.LockTypeNone,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			isErr: true,
			err:   repository.ErrRecordNotFound,
		},
		{
			description: "別のゲームのバージョンが混ざることはない",
			gameID:      gameID6,
			lockType:    repository.LockTypeNone,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID5),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
						},
					},
				},
				{
					ID:          uuid.UUID(gameID7),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID6),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
						},
					},
				},
			},
			expect: domain.NewGameVersion(
				gameVersionID5,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				time.Now(),
			),
		},
		{
			description: "lockTypeがRecordでもエラーなし",
			gameID:      gameID8,
			lockType:    repository.LockTypeRecord,
			games: []GameTable{
				{
					ID:          uuid.UUID(gameID8),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []GameVersionTable{
						{
							ID:          uuid.UUID(gameVersionID7),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
						},
					},
				},
			},
			expect: domain.NewGameVersion(
				gameVersionID7,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				time.Now(),
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.games) != 0 {
				err := db.Create(&testCase.games).Error
				if err != nil {
					t.Fatalf("failed to create games: %v", err)
				}

				for _, game := range testCase.games {
					if game.DeletedAt.Valid {
						err := db.Delete(&game).Error
						if err != nil {
							t.Fatalf("failed to delete game: %v", err)
						}
					}
				}
			}

			gameVersion, err := gameVersionRepository.GetLatestGameVersion(ctx, testCase.gameID, testCase.lockType)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Log(gameVersion)
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, testCase.expect.GetID(), gameVersion.GetID())
			assert.Equal(t, testCase.expect.GetName(), gameVersion.GetName())
			assert.Equal(t, testCase.expect.GetDescription(), gameVersion.GetDescription())
			assert.WithinDuration(t, testCase.expect.GetCreatedAt(), gameVersion.GetCreatedAt(), 2*time.Second)
		})
	}
}
