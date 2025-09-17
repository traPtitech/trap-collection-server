package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSaveEdition(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	editionRepository := NewEdition(testDB)

	type test struct {
		description     string
		edition         *domain.LauncherVersion
		beforeEditions  []migrate.EditionTable2
		noQuestionnaire bool
		isErr           bool
		err             error
	}

	now := time.Now()
	strURLLink := "https://example.com"
	urlLink, err := url.Parse(strURLLink)
	if err != nil {
		t.Fatalf("failed to encode url: %v", err)
	}
	//questionnaireURL := values.NewLauncherVersionQuestionnaireURL(urlLink)

	editionID := values.NewLauncherVersionID()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			edition: domain.NewLauncherVersionWithQuestionnaire(
				values.NewLauncherVersionID(),
				"test1",
				urlLink,
				now,
			),
		},
		{
			description: "別のバージョンが存在してもエラーなし",
			edition: domain.NewLauncherVersionWithQuestionnaire(
				values.NewLauncherVersionID(),
				"test2",
				urlLink,
				now,
			),
			beforeEditions: []migrate.EditionTable2{
				{
					ID:               uuid.New(),
					Name:             "test3",
					QuestionnaireURL: sql.NullString{String: urlLink.String(), Valid: true},
					CreatedAt:        now,
				},
			},
		},
		{
			description: "アンケートが無くてもエラーなし",
			edition: domain.NewLauncherVersionWithoutQuestionnaire(
				values.NewLauncherVersionID(),
				"test4",
				now,
			),
			noQuestionnaire: true,
		},
		{
			description: "同じバージョンIDが存在するのでエラー",
			edition: domain.NewLauncherVersionWithQuestionnaire(
				editionID,
				"test5",
				urlLink,
				now,
			),
			beforeEditions: []migrate.EditionTable2{
				{
					ID:               uuid.UUID(editionID),
					Name:             "test6",
					QuestionnaireURL: sql.NullString{String: urlLink.String(), Valid: true},
					CreatedAt:        now,
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.beforeEditions != nil {
				err := db.
					Session(&gorm.Session{}).
					Create(testCase.beforeEditions).Error
				if err != nil {
					t.Fatalf("failed to create edition: %+v\n", err)
				}
			}
		})

		err := editionRepository.SaveEdition(ctx, testCase.edition)
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

		var edition migrate.EditionTable2
		err = db.
			Session(&gorm.Session{}).
			Where("id = ?", uuid.UUID(testCase.edition.GetID())).
			Find(&edition).Error
		if err != nil {
			t.Fatalf("failed to get game files: %+v\n", err)
		}

		assert.Equal(t, uuid.UUID(testCase.edition.GetID()), edition.ID)
		assert.Equal(t, string(testCase.edition.GetName()), edition.Name)
		if !testCase.noQuestionnaire {
			assert.Equal(t, strURLLink, edition.QuestionnaireURL.String)
		}
		assert.WithinDuration(t, testCase.edition.GetCreatedAt(), edition.CreatedAt, time.Second)
	}
}

func TestUpdateEdition(t *testing.T) {

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	editionRepository := NewEdition(testDB)

	now := time.Now()
	urlLink1, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode url: %v", err)
	}

	urlLink2, err := url.Parse("https://example2.com")
	if err != nil {
		t.Fatalf("failed to encode url: %v", err)
	}

	editionID1 := values.NewLauncherVersionID()
	editionID2 := values.NewLauncherVersionID()

	type test struct {
		description    string
		edition        *domain.LauncherVersion
		beforeEditions []migrate.EditionTable2
		afterEditions  []migrate.EditionTable2
		isErr          bool
		err            error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			edition: domain.NewLauncherVersionWithQuestionnaire(
				editionID1,
				"test2",
				urlLink2,
				now,
			),
			beforeEditions: []migrate.EditionTable2{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test1",
					QuestionnaireURL: sql.NullString{
						String: urlLink1.String(),
						Valid:  true,
					},
					CreatedAt: now,
				},
			},
			afterEditions: []migrate.EditionTable2{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test2",
					QuestionnaireURL: sql.NullString{
						String: urlLink2.String(),
						Valid:  true,
					},
					CreatedAt: now,
				},
			},
		},
		{
			description: "別のエディションが存在してもエラーなし",
			edition: domain.NewLauncherVersionWithQuestionnaire(
				editionID1,
				"test3",
				urlLink2,
				now,
			),
			beforeEditions: []migrate.EditionTable2{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test1",
					QuestionnaireURL: sql.NullString{
						String: urlLink1.String(),
						Valid:  true,
					},
					CreatedAt: now,
				},
				{
					ID:   uuid.UUID(editionID2),
					Name: "test2",
					QuestionnaireURL: sql.NullString{
						String: urlLink1.String(),
						Valid:  true,
					},
					CreatedAt: now.Add(-time.Hour),
				},
			},
			afterEditions: []migrate.EditionTable2{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test3",
					QuestionnaireURL: sql.NullString{
						String: urlLink2.String(),
						Valid:  true,
					},
					CreatedAt: now,
				},
				{
					ID:   uuid.UUID(editionID2),
					Name: "test2",
					QuestionnaireURL: sql.NullString{
						String: urlLink1.String(),
						Valid:  true,
					},
					CreatedAt: now.Add(-time.Hour),
				},
			},
		},
		{
			description: "アンケートURLが存在しなくなってもエラーなし",
			edition: domain.NewLauncherVersionWithoutQuestionnaire(
				editionID1,
				"test2",
				now,
			),
			beforeEditions: []migrate.EditionTable2{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test1",
					QuestionnaireURL: sql.NullString{
						String: urlLink1.String(),
						Valid:  true,
					},
					CreatedAt: now,
				},
			},
			afterEditions: []migrate.EditionTable2{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test2",
					QuestionnaireURL: sql.NullString{
						Valid: false,
					},
					CreatedAt: now,
				},
			},
		},
		{
			description: "アンケートURLが存在するようになってもエラーなし",
			edition: domain.NewLauncherVersionWithQuestionnaire(
				editionID1,
				"test2",
				urlLink1,
				now,
			),
			beforeEditions: []migrate.EditionTable2{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test1",
					QuestionnaireURL: sql.NullString{
						Valid: false,
					},
					CreatedAt: now,
				},
			},
			afterEditions: []migrate.EditionTable2{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test2",
					QuestionnaireURL: sql.NullString{
						String: urlLink1.String(),
						Valid:  true,
					},
					CreatedAt: now,
				},
			},
		},
		{
			description: "アンケートURLが変わってもエラーなし",
			edition: domain.NewLauncherVersionWithQuestionnaire(
				editionID1,
				"test2",
				urlLink2,
				now,
			),
			beforeEditions: []migrate.EditionTable2{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test1",
					QuestionnaireURL: sql.NullString{
						String: urlLink1.String(),
						Valid:  true,
					},
					CreatedAt: now,
				},
			},
			afterEditions: []migrate.EditionTable2{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test2",
					QuestionnaireURL: sql.NullString{
						String: urlLink2.String(),
						Valid:  true,
					},
					CreatedAt: now,
				},
			},
		},
		{
			description: "エディションが無いのでErrNoRecordUpdated",
			edition: domain.NewLauncherVersionWithQuestionnaire(
				editionID1,
				"test2",
				urlLink1,
				now,
			),
			beforeEditions: []migrate.EditionTable2{},
			afterEditions:  []migrate.EditionTable2{},
			isErr:          true,
			err:            repository.ErrNoRecordUpdated,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Unscoped().
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Delete(&migrate.EditionTable2{}).Error
				if err != nil {
					t.Fatalf("failed to delete edition: %+v\n", err)
				}
			}()

			if len(testCase.beforeEditions) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Info),
					}).
					Create(&testCase.beforeEditions).Error
				if err != nil {
					t.Fatalf("failed to create edition: %+v\n", err)
				}
			}

			err := editionRepository.UpdateEdition(ctx, testCase.edition)

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

			var editions []migrate.EditionTable2
			err = db.
				Session(&gorm.Session{}).
				Order("created_at desc").
				Find(&editions).Error
			if err != nil {
				t.Fatalf("failed to get editions: %+v", err)
			}

			assert.Len(t, editions, len(testCase.afterEditions))
			for i, edition := range editions {
				assert.Equal(t, testCase.afterEditions[i].ID, edition.ID)
				assert.Equal(t, testCase.afterEditions[i].Name, edition.Name)
				assert.Equal(t, testCase.afterEditions[i].QuestionnaireURL.String, edition.QuestionnaireURL.String)
				assert.WithinDuration(t, testCase.afterEditions[i].CreatedAt, edition.CreatedAt, time.Second)
			}
		})
	}

}

func TestDeleteEdition(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	editionRepository := NewEdition(testDB)

	now := time.Now()
	urlLink1, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode url: %v", err)
	}

	urlLink2, err := url.Parse("https://example2.com")
	if err != nil {
		t.Fatalf("failed to encode url: %v", err)
	}

	editionID1 := values.NewLauncherVersionID()
	editionID2 := values.NewLauncherVersionID()

	type test struct {
		description    string
		editionID      values.LauncherVersionID
		beforeEditions []migrate.EditionTable2
		afterEditions  []migrate.EditionTable2
		isErr          bool
		err            error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラー無し",
			editionID:   editionID1,
			beforeEditions: []migrate.EditionTable2{
				{
					ID:               uuid.UUID(editionID1),
					Name:             "test1",
					QuestionnaireURL: sql.NullString{String: urlLink1.String()},
					CreatedAt:        now.Add(-time.Hour),
				},
			},
			afterEditions: []migrate.EditionTable2{},
		},
		{
			description: "他のゲームがあっても問題なし",
			editionID:   editionID1,
			beforeEditions: []migrate.EditionTable2{
				{
					ID:               uuid.UUID(editionID1),
					Name:             "test1",
					QuestionnaireURL: sql.NullString{String: urlLink1.String()},
					CreatedAt:        now.Add(-time.Hour),
				},
				{
					ID:               uuid.UUID(editionID2),
					Name:             "test2",
					QuestionnaireURL: sql.NullString{String: urlLink2.String(), Valid: true},
					CreatedAt:        now.Add(-time.Hour * 2),
				},
			},
			afterEditions: []migrate.EditionTable2{
				{
					ID:               uuid.UUID(editionID2),
					Name:             "test2",
					QuestionnaireURL: sql.NullString{String: urlLink2.String(), Valid: true},
					CreatedAt:        now.Add(-time.Hour * 2),
				},
			},
		},
		{
			description: "エディションが存在しないのでErrNoRecordDeleted",
			editionID:   editionID1,
			beforeEditions: []migrate.EditionTable2{
				{
					ID:               uuid.UUID(editionID2),
					Name:             "test1",
					QuestionnaireURL: sql.NullString{String: urlLink1.String()},
					CreatedAt:        now.Add(-time.Hour),
				},
			},
			isErr: true,
			err:   repository.ErrNoRecordDeleted,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Unscoped().
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Delete(&migrate.EditionTable2{}).Error
				if err != nil {
					t.Fatalf("failed to delete edition: %+v\n", err)
				}
			}()

			if len(testCase.beforeEditions) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Info),
					}).
					Create(&testCase.beforeEditions).Error
				if err != nil {
					t.Fatalf("failed to create edition: %+v\n", err)
				}
			}

			err := editionRepository.DeleteEdition(ctx, testCase.editionID)

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

			var editions []migrate.EditionTable2
			err = db.
				Session(&gorm.Session{}).
				Order("created_at desc").
				Find(&editions).Error
			if err != nil {
				t.Fatalf("failed to get editions: %+v", err)
			}

			assert.Len(t, editions, len(testCase.afterEditions))
			for i, edition := range editions {
				assert.Equal(t, testCase.afterEditions[i].ID, edition.ID)
				assert.Equal(t, testCase.afterEditions[i].Name, edition.Name)
				assert.Equal(t, testCase.afterEditions[i].QuestionnaireURL.String, edition.QuestionnaireURL.String)
				assert.WithinDuration(t, testCase.afterEditions[i].CreatedAt, edition.CreatedAt, time.Second)
			}
		})
	}
}

func TestUpdateEditionGameVersions(t *testing.T) {
	ctx := t.Context()

	db, err := testDB.getDB(ctx)
	require.NoError(t, err)

	editionID1 := values.NewLauncherVersionID()

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()

	gameID1 := values.NewGameID()

	gameImageID1 := values.NewGameImageID()
	gameVideoID1 := values.NewGameVideoID()

	// テスト用のゲーム、画像、動画を定義
	testGame := migrate.GameTable2{
		ID:                     uuid.UUID(gameID1),
		Name:                   "test game",
		Description:            "test description",
		CreatedAt:              time.Now(),
		VisibilityTypeID:       1,
		LatestVersionUpdatedAt: time.Now(),
	}
	testGameImage := migrate.GameImageTable2{
		ID:          uuid.UUID(gameImageID1),
		GameID:      uuid.UUID(gameID1),
		ImageTypeID: 1,
		CreatedAt:   time.Now(),
	}
	testGameVideo := migrate.GameVideoTable2{
		ID:          uuid.UUID(gameVideoID1),
		GameID:      uuid.UUID(gameID1),
		VideoTypeID: 1,
		CreatedAt:   time.Now(),
	}

	// テスト全体の開始前にデータを作成し、終了後に削除
	err = db.Create(&testGame).Error
	require.NoError(t, err)
	err = db.Create(&testGameImage).Error
	require.NoError(t, err)
	err = db.Create(&testGameVideo).Error
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx := context.Background() //この時点でtは終了しているので新しいコンテキストを作成

		err := db.WithContext(cleanupCtx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&migrate.GameVersionTable2{}).Error
		require.NoError(t, err)
		err = db.WithContext(cleanupCtx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&migrate.GameImageTable2{}).Error
		require.NoError(t, err)
		err = db.WithContext(cleanupCtx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&migrate.GameVideoTable2{}).Error
		require.NoError(t, err)
		err = db.WithContext(cleanupCtx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&migrate.GameTable2{}).Error
		require.NoError(t, err)
	})

	//テスト用の構造体
	type test struct {
		description    string
		editionID      values.LauncherVersionID
		gameVersionIDs []values.GameVersionID
		beforeEditions []migrate.EditionTable2
		afterEditions  []migrate.EditionTable2
		isErr          bool
		err            error
	}
	//リポジトリの作成
	editionRepository := NewEdition(testDB)

	//テストケースの書き出し　ゲームの作成もしていないとeditionが作成できないので注意
	testCases := []test{
		{
			description: "特に問題ないのでエラー無し",
			editionID:   editionID1,
			gameVersionIDs: []values.GameVersionID{
				gameVersionID1,
				gameVersionID2,
			},
			beforeEditions: []migrate.EditionTable2{
				{
					ID:        uuid.UUID(editionID1),
					Name:      "test1",
					CreatedAt: time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID1),
							Name:        "v1.0.0",
							GameID:      uuid.UUID(gameID1),
							CreatedAt:   time.Now(),
							GameImageID: uuid.UUID(gameImageID1),
							GameVideoID: uuid.UUID(gameVideoID1),
							Description: "test description",
						},
						{
							ID:          uuid.UUID(gameVersionID2),
							Name:        "v2.0.0",
							GameID:      uuid.UUID(gameID1),
							CreatedAt:   time.Now(),
							GameImageID: uuid.UUID(gameImageID1),
							GameVideoID: uuid.UUID(gameVideoID1),
							Description: "test description",
						},
					},
				},
			},
			afterEditions: []migrate.EditionTable2{
				{
					ID:        uuid.UUID(editionID1),
					Name:      "test1",
					CreatedAt: time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID1),
							Name:        "v1.0.0",
							GameID:      uuid.UUID(gameID1),
							CreatedAt:   time.Now(),
							GameImageID: uuid.UUID(gameImageID1),
							GameVideoID: uuid.UUID(gameVideoID1),
							Description: "test description",
						},
						{
							ID:          uuid.UUID(gameVersionID2),
							Name:        "v2.0.0",
							GameID:      uuid.UUID(gameID1),
							CreatedAt:   time.Now(),
							GameImageID: uuid.UUID(gameImageID1),
							GameVideoID: uuid.UUID(gameVideoID1),
							Description: "test description",
						},
					},
				},
			},
		},
		{
			description:    "空配列で全ての関連をクリア",
			editionID:      editionID1,
			gameVersionIDs: []values.GameVersionID{},
			beforeEditions: []migrate.EditionTable2{
				{
					ID:        uuid.UUID(editionID1),
					Name:      "test1",
					CreatedAt: time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID1),
							Name:        "v1.0.0",
							GameID:      uuid.UUID(gameID1),
							CreatedAt:   time.Now(),
							GameImageID: uuid.UUID(gameImageID1),
							GameVideoID: uuid.UUID(gameVideoID1),
							Description: "test description",
						},
					},
				},
			},
			afterEditions: []migrate.EditionTable2{
				{
					ID:           uuid.UUID(editionID1),
					Name:         "test1",
					CreatedAt:    time.Now(),
					GameVersions: []migrate.GameVersionTable2{},
				},
			},
		},
		{
			description:    "存在しないエディションIDでエラー",
			editionID:      values.NewLauncherVersionID(),
			gameVersionIDs: []values.GameVersionID{gameVersionID1},
			beforeEditions: []migrate.EditionTable2{
				{
					ID:        uuid.UUID(editionID1),
					Name:      "dummy edition",
					CreatedAt: time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID1),
							Name:        "v1.0.0",
							GameID:      uuid.UUID(gameID1),
							CreatedAt:   time.Now(),
							GameImageID: uuid.UUID(gameImageID1),
							GameVideoID: uuid.UUID(gameVideoID1),
							Description: "test description",
						},
					},
				},
			},
			afterEditions: []migrate.EditionTable2{
				{
					ID:        uuid.UUID(editionID1),
					Name:      "test1",
					CreatedAt: time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID1),
							Name:        "v1.0.0",
							GameID:      uuid.UUID(gameID1),
							CreatedAt:   time.Now(),
							GameImageID: uuid.UUID(gameImageID1),
							GameVideoID: uuid.UUID(gameVideoID1),
							Description: "test description",
						},
					},
				},
			},
			isErr: true,
			err:   nil,
		},
	}

	// テストケースの実行
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			// テストケースの削除
			t.Cleanup(func() {
				//テストで作成したEditionを取得
				var editions []migrate.EditionTable2
				err := db.Find(&editions).Error
				require.NoError(t, err)
				//各EditionのGameVersionsとの関連を解除
				for _, edition := range editions {
					err = db.Model(&edition).Association("GameVersions").Clear()
					require.NoError(t, err)
				}
				//親テーブルのデータを削除
				err = db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&migrate.EditionTable2{}).Error
				require.NoError(t, err)
			})

			if len(testCase.beforeEditions) != 0 {
				err = db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Silent),
					}).
					Create(&testCase.beforeEditions).Error
				require.NoError(t, err)
			}

			//テスト関数の実行
			err := editionRepository.UpdateEditionGameVersions(ctx, testCase.editionID, testCase.gameVersionIDs)

			//エラー処理
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

			var edition migrate.EditionTable2

			err = db.Where("id = ?", uuid.UUID(testCase.editionID)).
				Preload("GameVersions").
				First(&edition).Error
			require.NoError(t, err)

			var expectedGameVersionIDs []uuid.UUID
			if len(testCase.afterEditions) > 0 {
				expectedGameVersionIDs = make([]uuid.UUID, len(testCase.afterEditions[0].GameVersions))
				for i, gv := range testCase.afterEditions[0].GameVersions {
					expectedGameVersionIDs[i] = gv.ID
				}
			}

			actualIDs := make([]uuid.UUID, len(edition.GameVersions))
			for i, r := range edition.GameVersions {
				actualIDs[i] = r.ID
			}

			assert.Len(t, edition.GameVersions, len(testCase.afterEditions[0].GameVersions))
			if len(testCase.afterEditions[0].GameVersions) > 0 {
				assert.ElementsMatch(t, expectedGameVersionIDs, actualIDs)
			}
		})
	}
}
