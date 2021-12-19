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
		lockType          repository.LockType
		launcherVersion   *domain.LauncherVersion
		isErr             bool
		err               error
	}

	launcherVersionID1 := values.NewLauncherVersionID()
	launcherVersionID2 := values.NewLauncherVersionID()
	launcherVersionID3 := values.NewLauncherVersionID()

	questionnaireURL, err := url.Parse("https://example.com/questionnaire")
	if err != nil {
		t.Errorf("failed to create url: %v", err)
	}

	testCases := []test{
		{
			description:       "ランチャーバージョンが存在しないのでエラー",
			launcherVersionID: values.NewLauncherVersionID(),
			lockType:          repository.LockTypeNone,
			isErr:             true,
			err:               repository.ErrRecordNotFound,
		},
		{
			description:       "アンケートが存在しなくてもエラーなし",
			launcherVersionID: launcherVersionID1,
			lockType:          repository.LockTypeNone,
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				launcherVersionID1,
				"TestGetLauncherVersion1",
				time.Now(),
			),
		},
		{
			description:       "アンケートが存在してもエラーなし",
			launcherVersionID: launcherVersionID2,
			lockType:          repository.LockTypeNone,
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID2,
				"TestGetLauncherVersion2",
				values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
				time.Now(),
			),
		},
		{
			description:       "lockTypeがLockTypeRecordでもエラーなし",
			launcherVersionID: launcherVersionID3,
			lockType:          repository.LockTypeRecord,
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				launcherVersionID3,
				"TestGetLauncherVersion3",
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

			launcherVersion, err := launcherVersionRepository.GetLauncherVersion(ctx, testCase.launcherVersionID, testCase.lockType)

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

func TestGetLauncherVersionAndUserAndSessionByAccessToken(t *testing.T) {
	t.Parallel()

	launcherVersionRepository := NewLauncherVersion(testDB)

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

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

	productKey5, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	accessToken1, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	accessToken2, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	accessToken3, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	accessToken4, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	accessToken5, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	type test struct {
		description       string
		dbLauncherVersion LauncherVersionTable
		accessToken       values.LauncherSessionAccessToken
		launcherVersion   *domain.LauncherVersion
		launcherUser      *domain.LauncherUser
		launcherSession   *domain.LauncherSession
		isErr             bool
		err               error
	}

	launcherVersionID1 := values.NewLauncherVersionID()
	launcherVersionID2 := values.NewLauncherVersionID()
	launcherVersionID3 := values.NewLauncherVersionID()
	launcherVersionID4 := values.NewLauncherVersionID()
	launcherVersionID5 := values.NewLauncherVersionID()

	launcherUserID1 := values.NewLauncherUserID()
	launcherUserID2 := values.NewLauncherUserID()
	launcherUserID3 := values.NewLauncherUserID()
	launcherUserID5 := values.NewLauncherUserID()

	launcherSessionID1 := values.NewLauncherSessionID()
	launcherSessionID2 := values.NewLauncherSessionID()
	launcherSessionID3 := values.NewLauncherSessionID()
	launcherSessionID5 := values.NewLauncherSessionID()

	questionnaireURL, err := url.Parse("https://example.com/questionnaire")
	if err != nil {
		t.Errorf("failed to create url: %v", err)
	}

	now := time.Now()

	testCases := []test{
		{
			description: "通常の状態なので問題なし",
			accessToken: accessToken1,
			dbLauncherVersion: LauncherVersionTable{
				ID:        uuid.UUID(launcherVersionID1),
				Name:      "TestGetVersion,User,Session1",
				CreatedAt: now,
				LauncherUsers: []LauncherUserTable{
					{
						ID:         uuid.UUID(launcherUserID1),
						ProductKey: string(productKey1),
						CreatedAt:  now,
						LauncherSessions: []LauncherSessionTable{
							{
								ID:          uuid.UUID(launcherSessionID1),
								AccessToken: string(accessToken1),
								ExpiresAt:   now.Add(time.Hour),
								CreatedAt:   now,
							},
						},
					},
				},
			},
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				launcherVersionID1,
				values.NewLauncherVersionName("TestGetVersion,User,Session1"),
				now,
			),
			launcherUser: domain.NewLauncherUser(
				launcherUserID1,
				productKey1,
			),
			launcherSession: domain.NewLauncherSession(
				launcherSessionID1,
				accessToken1,
				now.Add(time.Hour),
			),
		},
		{
			description: "questionnaireURLが存在しても問題なし",
			accessToken: accessToken5,
			dbLauncherVersion: LauncherVersionTable{
				ID:   uuid.UUID(launcherVersionID5),
				Name: "TestGetVersion,User,Session5",
				QuestionnaireURL: sql.NullString{
					String: "https://example.com/questionnaire",
					Valid:  true,
				},
				CreatedAt: now,
				LauncherUsers: []LauncherUserTable{
					{
						ID:         uuid.UUID(launcherUserID5),
						ProductKey: string(productKey5),
						CreatedAt:  now,
						LauncherSessions: []LauncherSessionTable{
							{
								ID:          uuid.UUID(launcherSessionID5),
								AccessToken: string(accessToken5),
								ExpiresAt:   now.Add(time.Hour),
								CreatedAt:   now,
							},
						},
					},
				},
			},
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID5,
				values.NewLauncherVersionName("TestGetVersion,User,Session5"),
				values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
				now,
			),
			launcherUser: domain.NewLauncherUser(
				launcherUserID5,
				productKey5,
			),
			launcherSession: domain.NewLauncherSession(
				launcherSessionID5,
				accessToken5,
				now.Add(time.Hour),
			),
		},
		{
			description: "バージョンが削除されているのでエラー",
			accessToken: accessToken2,
			dbLauncherVersion: LauncherVersionTable{
				ID:        uuid.UUID(launcherVersionID2),
				Name:      "TestGetVersion,User,Session2",
				CreatedAt: now,
				DeletedAt: gorm.DeletedAt{
					Time:  now,
					Valid: true,
				},
				LauncherUsers: []LauncherUserTable{
					{
						ID:         uuid.UUID(launcherUserID2),
						ProductKey: string(productKey2),
						CreatedAt:  now,
						LauncherSessions: []LauncherSessionTable{
							{
								ID:          uuid.UUID(launcherSessionID2),
								AccessToken: string(accessToken2),
								ExpiresAt:   now.Add(time.Hour),
								CreatedAt:   now,
							},
						},
					},
				},
			},
			isErr: true,
			err:   repository.ErrRecordNotFound,
		},
		{
			description: "ユーザーが削除されているのでエラー",
			accessToken: accessToken3,
			dbLauncherVersion: LauncherVersionTable{
				ID:        uuid.UUID(launcherVersionID3),
				Name:      "TestGetVersion,User,Session3",
				CreatedAt: now,
				LauncherUsers: []LauncherUserTable{
					{
						ID:         uuid.UUID(launcherUserID3),
						ProductKey: string(productKey3),
						CreatedAt:  now,
						DeletedAt: gorm.DeletedAt{
							Time:  now,
							Valid: true,
						},
						LauncherSessions: []LauncherSessionTable{
							{
								ID:          uuid.UUID(launcherSessionID3),
								AccessToken: string(accessToken3),
								ExpiresAt:   now.Add(time.Hour),
								CreatedAt:   now,
							},
						},
					},
				},
			},
			isErr: true,
			err:   repository.ErrRecordNotFound,
		},
		{
			description: "アクセストークンが存在しないのでエラー",
			accessToken: accessToken4,
			dbLauncherVersion: LauncherVersionTable{
				ID:        uuid.UUID(launcherVersionID4),
				Name:      "TestGetVersion,User,Session4",
				CreatedAt: now,
			},
			isErr: true,
			err:   repository.ErrRecordNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&testCase.dbLauncherVersion).Error
			if err != nil {
				t.Errorf("failed to create test data: %v", err)
			}

			if testCase.dbLauncherVersion.DeletedAt.Valid {
				err = db.Delete(&testCase.dbLauncherVersion).Error
				if err != nil {
					t.Errorf("failed to delete test data: %v", err)
				}
			}
			for _, launcherUser := range testCase.dbLauncherVersion.LauncherUsers {
				if launcherUser.DeletedAt.Valid {
					err = db.Delete(&launcherUser).Error
					if err != nil {
						t.Errorf("failed to delete test data: %v", err)
					}
				}
				for _, launcherSession := range launcherUser.LauncherSessions {
					if launcherSession.DeletedAt.Valid {
						err = db.Delete(&launcherSession).Error
						if err != nil {
							t.Errorf("failed to delete test data: %v", err)
						}
					}
				}
			}

			launcherVersion, launcherUser, launcherSession, err := launcherVersionRepository.GetLauncherVersionAndUserAndSessionByAccessToken(ctx, testCase.accessToken)

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

			assert.Equal(t, testCase.launcherUser.GetID(), launcherUser.GetID())
			assert.Equal(t, testCase.launcherUser.GetProductKey(), launcherUser.GetProductKey())

			assert.Equal(t, testCase.launcherSession.GetID(), launcherSession.GetID())
			assert.Equal(t, testCase.launcherSession.GetAccessToken(), launcherSession.GetAccessToken())
			assert.WithinDuration(t, testCase.launcherSession.GetExpiresAt(), launcherSession.GetExpiresAt(), time.Second)
		})
	}
}
