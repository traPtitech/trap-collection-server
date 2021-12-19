package gorm2

import (
	"context"
	"errors"
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

func TestSaveGameURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameURLRepository := NewGameURL(testDB)

	type test struct {
		description        string
		gameVersionID      values.GameVersionID
		gameURL            *domain.GameURL
		beforeGameVersions []GameVersionTable
		afterGameURLs      []GameURLTable
		isErr              bool
		err                error
	}

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()
	gameVersionID7 := values.NewGameVersionID()

	gameURLID1 := values.NewGameURLID()
	gameURLID2 := values.NewGameURLID()
	gameURLID3 := values.NewGameURLID()
	gameURLID4 := values.NewGameURLID()
	gameURLID5 := values.NewGameURLID()
	gameURLID6 := values.NewGameURLID()
	gameURLID7 := values.NewGameURLID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}
	link := values.NewGameURLLink(urlLink)

	emptyLink := url.URL{}

	testCases := []test{
		{
			description:   "特に問題ないのでエラーなし",
			gameVersionID: gameVersionID1,
			gameURL:       domain.NewGameURL(gameURLID1, link, time.Now()),
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			afterGameURLs: []GameURLTable{
				{
					ID:            uuid.UUID(gameURLID1),
					GameVersionID: uuid.UUID(gameVersionID1),
					URL:           urlLink.String(),
				},
			},
		},
		{
			description:   "空のURLでもエラーなし",
			gameVersionID: gameVersionID2,
			gameURL:       domain.NewGameURL(gameURLID2, &emptyLink, time.Now()),
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID2),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			afterGameURLs: []GameURLTable{
				{
					ID:            uuid.UUID(gameURLID2),
					GameVersionID: uuid.UUID(gameVersionID2),
					URL:           emptyLink.String(),
				},
			},
		},
		{
			description:   "URLが既に存在するのでエラー",
			gameVersionID: gameVersionID3,
			gameURL:       domain.NewGameURL(gameURLID3, link, time.Now()),
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameURL: GameURLTable{
						ID:  uuid.UUID(gameURLID5),
						URL: urlLink.String(),
					},
				},
			},
			afterGameURLs: []GameURLTable{
				{
					ID:            uuid.UUID(gameURLID5),
					GameVersionID: uuid.UUID(gameVersionID3),
					URL:           urlLink.String(),
				},
			},
			isErr: true,
		},
		{
			description:   "他のゲームバージョンが存在してもエラーなし",
			gameVersionID: gameVersionID4,
			gameURL:       domain.NewGameURL(gameURLID4, link, time.Now()),
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
				{
					ID:          uuid.UUID(gameVersionID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now().Add(-time.Hour),
				},
			},
			afterGameURLs: []GameURLTable{
				{
					ID:            uuid.UUID(gameURLID4),
					GameVersionID: uuid.UUID(gameVersionID4),
					URL:           urlLink.String(),
				},
			},
		},
		{
			description:   "他のゲームバージョンにURLが存在してもエラーなし",
			gameVersionID: gameVersionID6,
			gameURL:       domain.NewGameURL(gameURLID7, link, time.Now()),
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
				{
					ID:          uuid.UUID(gameVersionID7),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now().Add(-time.Hour),
					GameURL: GameURLTable{
						ID:  uuid.UUID(gameURLID6),
						URL: urlLink.String(),
					},
				},
			},
			afterGameURLs: []GameURLTable{
				{
					ID:            uuid.UUID(gameURLID7),
					GameVersionID: uuid.UUID(gameVersionID6),
					URL:           urlLink.String(),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&GameTable{
				ID:           uuid.UUID(values.NewGameID()),
				Name:         "test",
				Description:  "test",
				CreatedAt:    time.Now(),
				GameVersions: testCase.beforeGameVersions,
			}).Error
			if err != nil {
				t.Fatalf("failed to create game table: %+v\n", err)
			}

			err = gameURLRepository.SaveGameURL(ctx, testCase.gameVersionID, testCase.gameURL)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			var urls []GameURLTable
			err = db.
				Session(&gorm.Session{}).
				Where("game_version_id = ?", uuid.UUID(testCase.gameVersionID)).
				Find(&urls).Error
			if err != nil {
				t.Fatalf("failed to get game url table: %+v\n", err)
			}

			assert.Len(t, urls, len(testCase.afterGameURLs))

			urlMap := make(map[uuid.UUID]GameURLTable)
			for _, url := range urls {
				urlMap[url.ID] = url
			}

			for _, expectURL := range testCase.afterGameURLs {
				actualURL, ok := urlMap[expectURL.ID]
				if !ok {
					t.Fatalf("not found url: %v", expectURL.ID)
				}

				assert.Equal(t, expectURL.GameVersionID, actualURL.GameVersionID)
				assert.Equal(t, expectURL.URL, actualURL.URL)
			}
		})
	}
}

func TestGetGameURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameURLRepository := NewGameURL(testDB)

	type test struct {
		description        string
		gameVersionID      values.GameVersionID
		beforeGameVersions []GameVersionTable
		gameURL            *domain.GameURL
		isErr              bool
		err                error
	}

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()
	gameVersionID7 := values.NewGameVersionID()
	gameVersionID8 := values.NewGameVersionID()
	gameVersionID9 := values.NewGameVersionID()

	gameURLID1 := values.NewGameURLID()
	gameURLID2 := values.NewGameURLID()
	gameURLID3 := values.NewGameURLID()
	gameURLID4 := values.NewGameURLID()
	gameURLID5 := values.NewGameURLID()
	gameURLID6 := values.NewGameURLID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}
	link := values.NewGameURLLink(urlLink)

	emptyLink := url.URL{}

	testCases := []test{
		{
			description:   "特に問題ないのでエラーなし",
			gameVersionID: gameVersionID1,
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameURL: GameURLTable{
						ID:            uuid.UUID(gameURLID1),
						GameVersionID: uuid.UUID(gameVersionID1),
						URL:           urlLink.String(),
					},
				},
			},
			gameURL: domain.NewGameURL(gameURLID1, link, time.Now()),
		},
		{
			description:   "空のURLでもエラーなし",
			gameVersionID: gameVersionID2,
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID2),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameURL: GameURLTable{
						ID:            uuid.UUID(gameURLID2),
						GameVersionID: uuid.UUID(gameVersionID2),
						URL:           emptyLink.String(),
					},
				},
			},
			gameURL: domain.NewGameURL(gameURLID2, &emptyLink, time.Now()),
		},
		{
			description:   "URLが存在しないのでErrRecordNotFound",
			gameVersionID: gameVersionID3,
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			isErr: true,
			err:   repository.ErrRecordNotFound,
		},
		{
			description:   "gameVersionが存在しないのでErrRecordNotFound",
			gameVersionID: gameVersionID4,
			isErr:         true,
			err:           repository.ErrRecordNotFound,
		},
		{
			description:   "他のゲームバージョンが存在してもエラーなし",
			gameVersionID: gameVersionID5,
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameURL: GameURLTable{
						ID:            uuid.UUID(gameURLID3),
						GameVersionID: uuid.UUID(gameVersionID5),
						URL:           urlLink.String(),
					},
				},
				{
					ID:          uuid.UUID(gameVersionID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now().Add(-time.Hour),
				},
			},
			gameURL: domain.NewGameURL(gameURLID3, link, time.Now()),
		},
		{
			description:   "他のゲームバージョンにURLが存在してもエラーなし",
			gameVersionID: gameVersionID7,
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID7),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameURL: GameURLTable{
						ID:            uuid.UUID(gameURLID5),
						GameVersionID: uuid.UUID(gameVersionID7),
						URL:           urlLink.String(),
					},
				},
				{
					ID:          uuid.UUID(gameVersionID8),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now().Add(-time.Hour),
					GameURL: GameURLTable{
						ID:  uuid.UUID(gameURLID4),
						URL: urlLink.String(),
					},
				},
			},
			gameURL: domain.NewGameURL(gameURLID5, link, time.Now()),
		},
		{
			description:   "誤った形式のURLがDBに入っていた場合、エラー",
			gameVersionID: gameVersionID9,
			beforeGameVersions: []GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID9),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameURL: GameURLTable{
						ID:            uuid.UUID(gameURLID6),
						GameVersionID: uuid.UUID(gameVersionID9),
						URL:           " http://foo.com",
					},
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&GameTable{
				ID:           uuid.UUID(values.NewGameID()),
				Name:         "test",
				Description:  "test",
				CreatedAt:    time.Now(),
				GameVersions: testCase.beforeGameVersions,
			}).Error
			if err != nil {
				t.Fatalf("failed to create game table: %+v\n", err)
			}

			gameURL, err := gameURLRepository.GetGameURL(ctx, testCase.gameVersionID)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if testCase.isErr {
				return
			}

			assert.Equal(t, testCase.gameURL.GetID(), gameURL.GetID())
			assert.Equal(t, testCase.gameURL.GetLink(), gameURL.GetLink())
		})
	}
}
