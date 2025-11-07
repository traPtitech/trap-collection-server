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
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
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
		beforeEditions  []schema.EditionTable
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
			beforeEditions: []schema.EditionTable{
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
			beforeEditions: []schema.EditionTable{
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

		var edition schema.EditionTable
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
		beforeEditions []schema.EditionTable
		afterEditions  []schema.EditionTable
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
			beforeEditions: []schema.EditionTable{
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
			afterEditions: []schema.EditionTable{
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
			beforeEditions: []schema.EditionTable{
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
			afterEditions: []schema.EditionTable{
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
			beforeEditions: []schema.EditionTable{
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
			afterEditions: []schema.EditionTable{
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
			beforeEditions: []schema.EditionTable{
				{
					ID:   uuid.UUID(editionID1),
					Name: "test1",
					QuestionnaireURL: sql.NullString{
						Valid: false,
					},
					CreatedAt: now,
				},
			},
			afterEditions: []schema.EditionTable{
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
			beforeEditions: []schema.EditionTable{
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
			afterEditions: []schema.EditionTable{
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
			beforeEditions: []schema.EditionTable{},
			afterEditions:  []schema.EditionTable{},
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
					Delete(&schema.EditionTable{}).Error
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

			var editions []schema.EditionTable
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
		beforeEditions []schema.EditionTable
		afterEditions  []schema.EditionTable
		isErr          bool
		err            error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラー無し",
			editionID:   editionID1,
			beforeEditions: []schema.EditionTable{
				{
					ID:               uuid.UUID(editionID1),
					Name:             "test1",
					QuestionnaireURL: sql.NullString{String: urlLink1.String()},
					CreatedAt:        now.Add(-time.Hour),
				},
			},
			afterEditions: []schema.EditionTable{},
		},
		{
			description: "他のゲームがあっても問題なし",
			editionID:   editionID1,
			beforeEditions: []schema.EditionTable{
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
			afterEditions: []schema.EditionTable{
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
			beforeEditions: []schema.EditionTable{
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
					Delete(&schema.EditionTable{}).Error
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

			var editions []schema.EditionTable
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
	editionID2 := values.NewLauncherVersionID()

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()

	gameID1 := values.NewGameID()

	gameImageID1 := values.NewGameImageID()
	gameVideoID1 := values.NewGameVideoID()

	// テスト用のゲーム、画像、動画、ゲームバージョンを定義
	testGame := schema.GameTable2{
		ID:                     uuid.UUID(gameID1),
		Name:                   "test game",
		Description:            "test description",
		CreatedAt:              time.Now(),
		VisibilityTypeID:       1,
		LatestVersionUpdatedAt: time.Now(),
	}
	testGameImage := schema.GameImageTable2{
		ID:          uuid.UUID(gameImageID1),
		GameID:      uuid.UUID(gameID1),
		ImageTypeID: 1,
		CreatedAt:   time.Now(),
	}
	testGameVideo := schema.GameVideoTable2{
		ID:          uuid.UUID(gameVideoID1),
		GameID:      uuid.UUID(gameID1),
		VideoTypeID: 1,
		CreatedAt:   time.Now(),
	}
	testGameVersion1 := schema.GameVersionTable2{
		ID:          uuid.UUID(gameVersionID1),
		Name:        "v1.0.0",
		GameID:      uuid.UUID(gameID1),
		CreatedAt:   time.Now(),
		GameImageID: uuid.UUID(gameImageID1),
		GameVideoID: uuid.UUID(gameVideoID1),
		Description: "test description",
	}
	testGameVersion2 := schema.GameVersionTable2{
		ID:          uuid.UUID(gameVersionID2),
		Name:        "v2.0.0",
		GameID:      uuid.UUID(gameID1),
		CreatedAt:   time.Now(),
		GameImageID: uuid.UUID(gameImageID1),
		GameVideoID: uuid.UUID(gameVideoID1),
		Description: "test description",
	}

	// テスト全体の開始前にデータを作成し、終了後に削除
	err = db.Create(&testGame).Error
	require.NoError(t, err)
	err = db.Create(&testGameImage).Error
	require.NoError(t, err)
	err = db.Create(&testGameVideo).Error
	require.NoError(t, err)
	err = db.Create(&testGameVersion1).Error
	require.NoError(t, err)
	err = db.Create(&testGameVersion2).Error
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx := context.Background() //この時点でtは終了しているので新しいコンテキストを作成

		err := db.WithContext(cleanupCtx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameVersionTable2{}).Error
		require.NoError(t, err)
		err = db.WithContext(cleanupCtx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameImageTable2{}).Error
		require.NoError(t, err)
		err = db.WithContext(cleanupCtx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameVideoTable2{}).Error
		require.NoError(t, err)
		err = db.WithContext(cleanupCtx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameTable2{}).Error
		require.NoError(t, err)
	})

	//テスト用の構造体
	type test struct {
		description    string
		editionID      values.LauncherVersionID
		gameVersionIDs []values.GameVersionID
		beforeEditions []schema.EditionTable
		afterEditions  []schema.EditionTable
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
			beforeEditions: []schema.EditionTable{
				{
					ID:           uuid.UUID(editionID1),
					Name:         "test1",
					CreatedAt:    time.Now(),
					GameVersions: []schema.GameVersionTable2{testGameVersion1},
				},
			},
			afterEditions: []schema.EditionTable{
				{
					ID:           uuid.UUID(editionID1),
					Name:         "test1",
					CreatedAt:    time.Now(),
					GameVersions: []schema.GameVersionTable2{testGameVersion1, testGameVersion2},
				},
			},
		},
		{
			description:    "空配列で全ての関連をクリア",
			editionID:      editionID1,
			gameVersionIDs: []values.GameVersionID{},
			beforeEditions: []schema.EditionTable{
				{
					ID:           uuid.UUID(editionID1),
					Name:         "test1",
					CreatedAt:    time.Now(),
					GameVersions: []schema.GameVersionTable2{testGameVersion1},
				},
			},
			afterEditions: []schema.EditionTable{
				{
					ID:           uuid.UUID(editionID1),
					Name:         "test1",
					CreatedAt:    time.Now(),
					GameVersions: []schema.GameVersionTable2{},
				},
			},
		},
		{
			description:    "存在しないエディションIDでエラー",
			editionID:      values.NewLauncherVersionID(),
			gameVersionIDs: []values.GameVersionID{gameVersionID1},
			beforeEditions: []schema.EditionTable{
				{
					ID:           uuid.UUID(editionID1),
					Name:         "dummy edition",
					CreatedAt:    time.Now(),
					GameVersions: []schema.GameVersionTable2{testGameVersion1},
				},
			},
			afterEditions: []schema.EditionTable{
				{
					ID:           uuid.UUID(editionID1),
					Name:         "test1",
					CreatedAt:    time.Now(),
					GameVersions: []schema.GameVersionTable2{testGameVersion1},
				},
			},
			isErr: true,
			err:   nil,
		},
		{
			description: "editionが複数あっても問題なし", //edition1 2のうち、edition1だけ更新されることを確認
			editionID:   editionID1,
			gameVersionIDs: []values.GameVersionID{
				gameVersionID1,
				gameVersionID2,
			},
			beforeEditions: []schema.EditionTable{
				{
					ID:           uuid.UUID(editionID1),
					Name:         "test1",
					CreatedAt:    time.Now(),
					GameVersions: []schema.GameVersionTable2{testGameVersion1},
				},
				{
					ID:           uuid.UUID(editionID2),
					Name:         "test2",
					CreatedAt:    time.Now(),
					GameVersions: []schema.GameVersionTable2{testGameVersion1},
				},
			},
			afterEditions: []schema.EditionTable{
				{
					ID:           uuid.UUID(editionID1),
					Name:         "test1",
					CreatedAt:    time.Now(),
					GameVersions: []schema.GameVersionTable2{testGameVersion1, testGameVersion2},
				},
				{
					ID:           uuid.UUID(editionID2),
					Name:         "test2",
					CreatedAt:    time.Now(),
					GameVersions: []schema.GameVersionTable2{testGameVersion1},
				},
			},
		},
	}

	// テストケースの実行
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			// テストケースの削除
			t.Cleanup(func() {
				//テストで作成したEditionを取得
				var editions []schema.EditionTable
				err := db.Find(&editions).Error
				require.NoError(t, err)
				//各EditionのGameVersionsとの関連を解除
				for _, edition := range editions {
					err = db.Model(&edition).Association("GameVersions").Clear()
					require.NoError(t, err)
				}
				//親テーブルのデータを削除
				err = db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.EditionTable{}).Error
				require.NoError(t, err)
			})

			if len(testCase.beforeEditions) != 0 {
				err = db.
					Session(&gorm.Session{}).
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

			//editionは複数あるケースを想定して、配列で定義
			var editions []schema.EditionTable
			err = db.Preload("GameVersions").Find(&editions).Error
			require.NoError(t, err)

			// afterEditionsをマップに
			expectedMap := map[uuid.UUID][]uuid.UUID{}
			for _, e := range testCase.afterEditions {
				ids := make([]uuid.UUID, len(e.GameVersions))
				for i, gv := range e.GameVersions {
					ids[i] = gv.ID
				}
				expectedMap[e.ID] = ids
			}

			for _, edition := range editions {
				expectedIDs, ok := expectedMap[edition.ID]
				assert.True(t, ok, "unexpected edition ID: %s", edition.ID)
				actualIDs := make([]uuid.UUID, len(edition.GameVersions))
				for i, gv := range edition.GameVersions {
					actualIDs[i] = gv.ID
				}
				assert.ElementsMatch(t, expectedIDs, actualIDs)
			}
		})
	}
}
