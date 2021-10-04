package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"gorm.io/gorm"
)

func TestGetLauncherVersion(t *testing.T) {
	t.Parallel()

	launcherVersionRepository := NewLauncherVersion(testDB)

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		description       string
		launcherVersionID values.LauncherVersionID
		launcherVersion   *domain.LauncherVersion
		isErr             bool
		err               error
	}

	launcherVersionID1 := values.NewLauncherVersionID()
	launcherVersionID2 := values.NewLauncherVersionID()

	questionnaireURL, err := url.Parse("https://example.com/questionnaire")
	if err != nil {
		t.Errorf("failed to create url: %v", err)
	}

	testCases := []test{
		{
			description:       "ランチャーバージョンが存在しないのでエラー",
			launcherVersionID: values.NewLauncherVersionID(),
			isErr:             true,
			err:               repository.ErrRecordNotFound,
		},
		{
			description:       "アンケートが存在しなくてもエラーなし",
			launcherVersionID: launcherVersionID1,
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				launcherVersionID1,
				"TestGetLauncherVersion1",
				time.Now(),
			),
		},
		{
			description:       "アンケートが存在してもエラーなし",
			launcherVersionID: launcherVersionID2,
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID2,
				"TestGetLauncherVersion2",
				values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
				time.Now(),
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.launcherVersion != nil {
				dbLauncherVersion := LauncherVersionTable{
					ID:        uuid.UUID(testCase.launcherVersion.GetID()),
					Name:      string(testCase.launcherVersion.GetName()),
					CreatedAt: testCase.launcherVersion.GetCreatedAt(),
				}

				questionnaireURL, err := testCase.launcherVersion.GetQuestionnaireURL()
				if err != nil && !errors.Is(err, domain.ErrNoQuestionnaire) {
					t.Errorf("failed to get questionnaire url: %v", err)
				}

				if errors.Is(err, domain.ErrNoQuestionnaire) {
					dbLauncherVersion.QuestionnaireURL = sql.NullString{
						Valid: false,
					}
				} else {
					dbLauncherVersion.QuestionnaireURL = sql.NullString{
						String: (*url.URL)(questionnaireURL).String(),
						Valid:  true,
					}
				}

				err = db.Create(&dbLauncherVersion).Error
				if err != nil {
					t.Fatalf("failed to create launcher version: %v", err)
				}
			}

			launcherVersion, err := launcherVersionRepository.GetLauncherVersion(ctx, testCase.launcherVersionID)

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

			assert.Equal(t, testCase.launcherVersion.GetID(), launcherVersion.GetID())
			assert.Equal(t, testCase.launcherVersion.GetName(), launcherVersion.GetName())

			expectQuestionnaireURL, _ := testCase.launcherVersion.GetQuestionnaireURL()
			actualQuestionnaireURL, _ := launcherVersion.GetQuestionnaireURL()
			assert.Equal(t, expectQuestionnaireURL, actualQuestionnaireURL)

			assert.WithinDuration(t, testCase.launcherVersion.GetCreatedAt(), launcherVersion.GetCreatedAt(), time.Second)
		})
	}
}

func TestGetLauncherUsersByLauncherVersionID(t *testing.T) {
	t.Parallel()

	launcherVersionRepository := NewLauncherVersion(testDB)

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		description     string
		dbLauncherUsers []LauncherUserTable
		launcherUsers   []*domain.LauncherUser
		isErr           bool
		err             error
	}

	launcherUserID1 := values.NewLauncherUserID()
	launcherUserID2 := values.NewLauncherUserID()
	launcherUserID3 := values.NewLauncherUserID()

	productKey1, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	productKey2, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	productKey3, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	testCases := []test{
		{
			description: "ユーザーが存在するのでエラーなし",
			dbLauncherUsers: []LauncherUserTable{
				{
					ID:         uuid.UUID(launcherUserID1),
					ProductKey: string(productKey1),
					CreatedAt:  time.Now(),
				},
			},
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					launcherUserID1,
					productKey1,
				),
			},
		},
		{
			description:     "ユーザーが存在しなくてもエラーなし",
			dbLauncherUsers: []LauncherUserTable{},
			launcherUsers:   []*domain.LauncherUser{},
		},
		{
			description: "ユーザーが複数人でもエラーなし",
			dbLauncherUsers: []LauncherUserTable{
				{
					ID:         uuid.UUID(launcherUserID2),
					ProductKey: string(productKey2),
					CreatedAt:  time.Now(),
				},
				{
					ID:         uuid.UUID(launcherUserID3),
					ProductKey: string(productKey3),
					CreatedAt:  time.Now(),
				},
			},
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					launcherUserID2,
					productKey2,
				),
				domain.NewLauncherUser(
					launcherUserID3,
					productKey3,
				),
			},
		},
		{
			description: "削除されたユーザーは含まれない",
			dbLauncherUsers: []LauncherUserTable{
				{
					ID:         uuid.UUID(launcherUserID1),
					ProductKey: string(productKey1),
					CreatedAt:  time.Now(),
					DeletedAt: gorm.DeletedAt{
						Time:  time.Now(),
						Valid: true,
					},
				},
			},
			launcherUsers: []*domain.LauncherUser{},
		},
	}

	for i, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.description, func(t *testing.T) {
			launcherVersionID := values.NewLauncherVersionID()
			dbLauncherVersion := LauncherVersionTable{
				ID:            uuid.UUID(launcherVersionID),
				Name:          fmt.Sprintf("TestCreateLauncherUsers%d", i),
				CreatedAt:     time.Now(),
				LauncherUsers: testCase.dbLauncherUsers,
			}
			err := db.Create(&dbLauncherVersion).Error
			if err != nil {
				t.Errorf("failed to create launcher version: %v", err)
			}

			deletedLauncherUserIDs := []uuid.UUID{}
			for _, dbLauncherUser := range testCase.dbLauncherUsers {
				if dbLauncherUser.DeletedAt.Valid {
					deletedLauncherUserIDs = append(deletedLauncherUserIDs, dbLauncherUser.ID)
				}
			}
			if len(deletedLauncherUserIDs) > 0 {
				err = db.
					Where("id IN ?", deletedLauncherUserIDs).
					Delete(&LauncherUserTable{}).Error
				if err != nil {
					t.Errorf("failed to delete launcher user: %v", err)
				}
			}

			launcherUsers, err := launcherVersionRepository.GetLauncherUsersByLauncherVersionID(ctx, values.NewLauncherVersionIDFromUUID(dbLauncherVersion.ID))

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

			assert.Equal(t, len(testCase.launcherUsers), len(launcherUsers))
			launcherUserMap := make(map[values.LauncherUserID]*domain.LauncherUser, len(launcherUsers))
			for _, launcherUser := range launcherUsers {
				launcherUserMap[launcherUser.GetID()] = launcherUser
			}
			for _, launcherUser := range testCase.launcherUsers {
				actualLauncherUser := launcherUserMap[launcherUser.GetID()]
				assert.Equal(t, launcherUser.GetID(), actualLauncherUser.GetID())
				assert.Equal(t, launcherUser.GetProductKey(), actualLauncherUser.GetProductKey())
			}
		})
	}
}
