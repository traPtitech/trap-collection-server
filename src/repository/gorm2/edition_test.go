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
		description    string
		edition        *domain.LauncherVersion
		beforeEditions []migrate.EditionTable2
		isErr          bool
		err            error
	}

	now := time.Now()
	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode url: %v", err)
	}
	questionnaireURL := values.NewLauncherVersionQuestionnaireURL(urlLink)

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
					QuestionnaireURL: sql.NullString{String: urlLink.String()},
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
					QuestionnaireURL: sql.NullString{String: urlLink.String()},
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
		savedURL, err := url.Parse(edition.QuestionnaireURL.String)
		if err != nil {
			t.Fatalf("failed to parse url: %+v_n", err)
		}
		assert.Equal(t, questionnaireURL, values.LauncherVersionQuestionnaireURL(savedURL))
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

			if testCase.beforeEditions != nil && len(testCase.beforeEditions) != 0 {
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

			if testCase.beforeEditions != nil && len(testCase.beforeEditions) != 0 {
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
