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
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}
	editionID1 := values.NewLauncherVersionID()

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()

	gameID1 := values.NewGameID()

	gameImageID1 := values.NewGameImageID()
	gameVideoID1 := values.NewGameVideoID()

	type test struct {
		description        string
		editionID          values.LauncherVersionID
		gameVersionIDs     []values.GameVersionID
		beforeGameVersions []migrate.GameVersionTable2
		beforeEditions     []migrate.EditionTable2
		afterGameVersions  []migrate.GameVersionTable2
		beforeRelations    [][2]uuid.UUID
		isErr              bool
		err                error
	}
	editionRepository := NewEdition(testDB)

	testCases := []test{
		{description: "特に問題ないのでエラー無し",
			editionID:      editionID1,
			gameVersionIDs: []values.GameVersionID{gameVersionID1, gameVersionID2},
			beforeEditions: []migrate.EditionTable2{
				{
					ID:        uuid.UUID(editionID1),
					Name:      "test1",
					CreatedAt: time.Now(),
				},
			},
			beforeRelations: [][2]uuid.UUID{
				{uuid.UUID(editionID1), uuid.UUID(gameVersionID1)},
			},
			beforeGameVersions: []migrate.GameVersionTable2{
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
					Name:        "v1.0.0",
					GameID:      uuid.UUID(gameID1),
					CreatedAt:   time.Now(),
					GameImageID: uuid.UUID(gameImageID1),
					GameVideoID: uuid.UUID(gameVideoID1),
					Description: "test description",
				},
			},
			afterGameVersions: []migrate.GameVersionTable2{
				{ID: uuid.UUID(gameVersionID1)},
				{ID: uuid.UUID(gameVersionID2)},
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
				},
			},
			beforeGameVersions: []migrate.GameVersionTable2{
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
			beforeRelations: [][2]uuid.UUID{
				{uuid.UUID(editionID1), uuid.UUID(gameVersionID1)},
				{uuid.UUID(editionID1), uuid.UUID(gameVersionID2)},
			},
			afterGameVersions: []migrate.GameVersionTable2{},
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
        },
    },
			beforeGameVersions: []migrate.GameVersionTable2{
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
			beforeRelations: [][2]uuid.UUID{},
			isErr:           true,
			err:             repository.ErrNoRecordUpdated,
		},
	}
	// テストケースの実行
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				// 1. 関連テーブルのクリーンアップ
				err := db.Exec("DELETE FROM edition_game_version_relations WHERE edition_id = ?",
					uuid.UUID(testCase.editionID)).Error
				if err != nil {
					t.Fatalf("failed to clean up relations: %+v\n", err)
				}

				// 2. ゲームバージョンのクリーンアップ（画像・動画より先に削除）
				if len(testCase.beforeGameVersions) > 0 {
					ids := make([]uuid.UUID, len(testCase.beforeGameVersions))
					for i, gv := range testCase.beforeGameVersions {
						ids[i] = gv.ID
					}
					err = db.Unscoped().Where("id IN ?", ids).Delete(&migrate.GameVersionTable2{}).Error
					if err != nil {
						t.Fatalf("failed to delete game versions: %+v\n", err)
					}
				}

				// 3. ゲーム画像・動画のクリーンアップ（ゲームバージョン削除後）
				if len(testCase.beforeGameVersions) > 0 {
					err = db.Unscoped().Where("id = ?", uuid.UUID(gameImageID1)).Delete(&migrate.GameImageTable2{}).Error
					if err != nil {
						t.Fatalf("failed to delete game image: %+v\n", err)
					}

					err = db.Unscoped().Where("id = ?", uuid.UUID(gameVideoID1)).Delete(&migrate.GameVideoTable2{}).Error
					if err != nil {
						t.Fatalf("failed to delete game video: %+v\n", err)
					}
				}

				// 4. ゲームのクリーンアップ
				err = db.Unscoped().Where("id = ?", uuid.UUID(gameID1)).Delete(&migrate.GameTable2{}).Error
				if err != nil {
					t.Fatalf("failed to delete game: %+v\n", err)
				}

				// 5. エディションのクリーンアップ
				if len(testCase.beforeEditions) > 0 {
					err = db.Unscoped().Where("id = ?", uuid.UUID(testCase.editionID)).Delete(&migrate.EditionTable2{}).Error
					if err != nil {
						t.Fatalf("failed to delete edition: %+v\n", err)
					}
				}
			}()

			if len(testCase.beforeEditions) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Silent),
					}).
					Create(&testCase.beforeEditions).Error
				if err != nil {
					t.Fatalf("failed to create edition: %+v\n", err)
				}
			}

			if len(testCase.beforeGameVersions) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Silent),
					}).
					Create(&migrate.GameTable2{
						ID:                     uuid.UUID(gameID1),
						Name:                   "Test Game",
						Description:            "Test game description",
						VisibilityTypeID:       1,
						CreatedAt:              time.Now(),
						LatestVersionUpdatedAt: time.Now(),
					}).Error
				if err != nil {
					t.Fatalf("failed to create game: %+v\n", err)
				}
				// ゲームバージョンの画像と動画も作成
				err = db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Silent),
					}).
					Create(&migrate.GameImageTable2{
						ID:          uuid.UUID(gameImageID1),
						GameID:      uuid.UUID(gameID1),
						ImageTypeID: 1,
						CreatedAt:   time.Now(),
					}).Error
				if err != nil {
					t.Fatalf("failed to create game image: %+v\n", err)
				}

				err = db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Silent),
					}).
					Create(&migrate.GameVideoTable2{
						ID:          uuid.UUID(gameVideoID1),
						GameID:      uuid.UUID(gameID1),
						VideoTypeID: 1,
						CreatedAt:   time.Now(),
					}).Error
				if err != nil {
					t.Fatalf("failed to create game video: %+v\n", err)
				}

				err = db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Silent),
					}).
					Create(&testCase.beforeGameVersions).Error
				if err != nil {
					t.Fatalf("failed to create game version: %+v\n", err)
				}
			}

			for _, relation := range testCase.beforeRelations {
				err := db.Exec(`
					INSERT INTO edition_game_version_relations (edition_id, game_version_id)
					VALUES (?, ?)
				`, relation[0], relation[1]).Error
				if err != nil {
					t.Fatalf("failed to create edition-game version relation: %+v\n", err)
				}
			}

			err := editionRepository.UpdateEditionGameVersions(ctx, testCase.editionID, testCase.gameVersionIDs)

			if testCase.isErr {
				if testCase.err != nil {
					if !errors.Is(err, testCase.err) {
						t.Fatalf("expected error %v, got %v", testCase.err, err)
					}
				} else {
					assert.Error(t, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			if err != nil {
				return
			}

			var relations []struct {
				GameVersionID uuid.UUID
			}
			err = db.Raw(`
				SELECT game_version_id 
				FROM edition_game_version_relations 
				WHERE edition_id = ?
			`, uuid.UUID(testCase.editionID)).Scan(&relations).Error
			if err != nil {
				t.Fatalf("failed to get edition-game version relations: %+v", err)
			}

			expectedIDs := make([]uuid.UUID, len(testCase.afterGameVersions))
			for i, gv := range testCase.afterGameVersions {
				expectedIDs[i] = gv.ID
			}

			actualIDs := make([]uuid.UUID, len(relations))
			for i, r := range relations {
				actualIDs[i] = r.GameVersionID
			}

			assert.Len(t, relations, len(testCase.afterGameVersions))
			if len(testCase.afterGameVersions) > 0 {
				assert.ElementsMatch(t, expectedIDs, actualIDs)
			}
		})
	}
}
