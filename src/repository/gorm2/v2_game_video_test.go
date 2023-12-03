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
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func TestSaveGameVideoV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameVideoRepository := NewGameVideoV2(testDB)

	type test struct {
		description  string
		gameID       values.GameID
		video        *domain.GameVideo
		beforeVideos []migrate.GameVideoTable2
		expectVideos []migrate.GameVideoTable2
		isErr        bool
		err          error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()

	videoID1 := values.NewGameVideoID()
	videoID2 := values.NewGameVideoID()
	videoID3 := values.NewGameVideoID()
	videoID4 := values.NewGameVideoID()
	videoID5 := values.NewGameVideoID()
	videoID6 := values.NewGameVideoID()

	var videoTypes []*migrate.GameVideoTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&videoTypes).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	videoTypeMap := make(map[string]int, len(videoTypes))
	for _, videoType := range videoTypes {
		videoTypeMap[videoType.Name] = videoType.ID
	}

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			gameID:      gameID1,
			video: domain.NewGameVideo(
				videoID1,
				values.GameVideoTypeMp4,
				now,
			),
			beforeVideos: []migrate.GameVideoTable2{},
			expectVideos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID1),
					GameID:      uuid.UUID(gameID1),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now,
				},
			},
		},
		{
			description: "想定外の動画の種類なのでエラー",
			gameID:      gameID2,
			video: domain.NewGameVideo(
				videoID2,
				100,
				now,
			),
			beforeVideos: []migrate.GameVideoTable2{},
			expectVideos: []migrate.GameVideoTable2{},
			isErr:        true,
		},
		{
			description: "既に動画が存在しても問題なし",
			gameID:      gameID3,
			video: domain.NewGameVideo(
				videoID3,
				values.GameVideoTypeMp4,
				now,
			),
			beforeVideos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID4),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectVideos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID4),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
				{
					ID:          uuid.UUID(videoID3),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now,
				},
			},
		},
		{
			description: "エラーの場合変更なし",
			gameID:      gameID4,
			video: domain.NewGameVideo(
				videoID5,
				100,
				now,
			),
			beforeVideos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID6),
					GameID:      uuid.UUID(gameID4),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectVideos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID6),
					GameID:      uuid.UUID(gameID4),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&migrate.GameTable2{
				ID:               uuid.UUID(testCase.gameID),
				Name:             "test",
				Description:      "test",
				CreatedAt:        time.Now(),
				GameVideo2s:      testCase.beforeVideos,
				VisibilityTypeID: gameVisibilityTypeIDPublic,
			}).Error
			if err != nil {
				t.Fatalf("failed to create game table: %+v\n", err)
			}

			err = gameVideoRepository.SaveGameVideo(ctx, testCase.gameID, testCase.video)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			var videos []migrate.GameVideoTable2
			err = db.
				Session(&gorm.Session{}).
				Where("game_id = ?", uuid.UUID(testCase.gameID)).
				Find(&videos).Error
			if err != nil {
				t.Fatalf("failed to get role table: %+v\n", err)
			}

			assert.Len(t, videos, len(testCase.expectVideos))

			videoMap := make(map[uuid.UUID]migrate.GameVideoTable2)
			for _, video := range videos {
				videoMap[video.ID] = video
			}

			for _, expectVideo := range testCase.expectVideos {
				actualVideo, ok := videoMap[expectVideo.ID]
				if !ok {
					t.Errorf("not found video: %+v", expectVideo)
				}

				assert.Equal(t, expectVideo.GameID, actualVideo.GameID)
				assert.Equal(t, expectVideo.VideoTypeID, actualVideo.VideoTypeID)
				assert.WithinDuration(t, expectVideo.CreatedAt, actualVideo.CreatedAt, 2*time.Second)
			}
		})
	}
}

func TestGetGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameVideoRepository := NewGameVideoV2(testDB)

	type test struct {
		description string
		videoID     values.GameVideoID
		lockType    repository.LockType
		videos      []migrate.GameVideoTable2
		expectVideo repository.GameVideoInfo
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()

	videoID1 := values.NewGameVideoID()
	videoID2 := values.NewGameVideoID()
	videoID3 := values.NewGameVideoID()
	videoID4 := values.NewGameVideoID()
	videoID5 := values.NewGameVideoID()

	var videoTypes []*migrate.GameVideoTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&videoTypes).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	videoTypeMap := make(map[string]int, len(videoTypes))
	for _, videoType := range videoTypes {
		videoTypeMap[videoType.Name] = videoType.ID
	}

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			videoID:     videoID1,
			lockType:    repository.LockTypeNone,
			videos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID1),
					GameID:      uuid.UUID(gameID1),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now,
				},
			},
			expectVideo: repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID1,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID1,
			},
		},
		{
			description: "lockTypeがRecordでも問題なし",
			videoID:     videoID2,
			lockType:    repository.LockTypeRecord,
			videos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID2),
					GameID:      uuid.UUID(gameID2),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now,
				},
			},
			expectVideo: repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID2,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID2,
			},
		},
		{
			description: "複数の動画があっても問題なし",
			videoID:     videoID3,
			videos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID3),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(videoID4),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectVideo: repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID3,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID3,
			},
		},
		{
			description: "動画が存在しないのでRecordNotFound",
			videoID:     videoID5,
			lockType:    repository.LockTypeNone,
			videos:      []migrate.GameVideoTable2{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			gameIDMap := map[uuid.UUID]*migrate.GameTable2{}
			for _, video := range testCase.videos {
				if game, ok := gameIDMap[video.GameID]; ok {
					game.GameVideo2s = append(game.GameVideo2s, video)
				} else {
					gameIDMap[video.GameID] = &migrate.GameTable2{
						ID:               video.GameID,
						Name:             "test",
						Description:      "test",
						CreatedAt:        now,
						GameVideo2s:      []migrate.GameVideoTable2{video},
						VisibilityTypeID: gameVisibilityTypeIDPublic,
					}
				}
			}

			games := make([]migrate.GameTable2, 0, len(gameIDMap))
			for _, game := range gameIDMap {
				games = append(games, *game)
			}

			if len(games) > 0 {
				err := db.Create(games).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}
			}

			video, err := gameVideoRepository.GetGameVideo(ctx, testCase.videoID, testCase.lockType)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, testCase.expectVideo.GameVideo.GetID(), video.GameVideo.GetID())
			assert.Equal(t, testCase.expectVideo.GameVideo.GetType(), video.GameVideo.GetType())
			assert.WithinDuration(t, testCase.expectVideo.GameVideo.GetCreatedAt(), video.GameVideo.GetCreatedAt(), time.Second)
			assert.Equal(t, testCase.expectVideo.GameID, video.GameID)
		})
	}
}

func TestGetGameVideos(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameVideoRepository := NewGameVideoV2(testDB)

	type test struct {
		description  string
		gameID       values.GameID
		lockType     repository.LockType
		videos       []migrate.GameVideoTable2
		expectVideos []*domain.GameVideo
		isErr        bool
		err          error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()

	videoID1 := values.NewGameVideoID()
	videoID2 := values.NewGameVideoID()
	videoID3 := values.NewGameVideoID()
	videoID4 := values.NewGameVideoID()

	var videoTypes []*migrate.GameVideoTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&videoTypes).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	videoTypeMap := make(map[string]int, len(videoTypes))
	for _, videoType := range videoTypes {
		videoTypeMap[videoType.Name] = videoType.ID
	}

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			gameID:      gameID1,
			lockType:    repository.LockTypeNone,
			videos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID1),
					GameID:      uuid.UUID(gameID1),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now,
				},
			},
			expectVideos: []*domain.GameVideo{
				domain.NewGameVideo(
					videoID1,
					values.GameVideoTypeMp4,
					now,
				),
			},
		},
		{
			description: "lockTypeがRecordでも問題なし",
			gameID:      gameID2,
			lockType:    repository.LockTypeRecord,
			videos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID2),
					GameID:      uuid.UUID(gameID2),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now,
				},
			},
			expectVideos: []*domain.GameVideo{
				domain.NewGameVideo(
					videoID2,
					values.GameVideoTypeMp4,
					now,
				),
			},
		},
		{
			description: "複数の動画があっても問題なし",
			gameID:      gameID3,
			videos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID3),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(videoID4),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[migrate.GameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectVideos: []*domain.GameVideo{
				domain.NewGameVideo(
					videoID3,
					values.GameVideoTypeMp4,
					now,
				),
				domain.NewGameVideo(
					videoID4,
					values.GameVideoTypeMp4,
					now.Add(-10*time.Hour),
				),
			},
		},
		{
			description:  "動画が存在しなくても問題なし",
			gameID:       gameID4,
			lockType:     repository.LockTypeNone,
			videos:       []migrate.GameVideoTable2{},
			expectVideos: []*domain.GameVideo{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			gameIDMap := map[uuid.UUID]*migrate.GameTable2{}
			for _, video := range testCase.videos {
				if game, ok := gameIDMap[video.GameID]; ok {
					game.GameVideo2s = append(game.GameVideo2s, video)
				} else {
					gameIDMap[video.GameID] = &migrate.GameTable2{
						ID:               video.GameID,
						Name:             "test",
						Description:      "test",
						CreatedAt:        now,
						GameVideo2s:      []migrate.GameVideoTable2{video},
						VisibilityTypeID: gameVisibilityTypeIDPublic,
					}
				}
			}

			games := make([]migrate.GameTable2, 0, len(gameIDMap))
			for _, game := range gameIDMap {
				games = append(games, *game)
			}

			if len(games) > 0 {
				err := db.Create(games).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}
			}

			videos, err := gameVideoRepository.GetGameVideos(ctx, testCase.gameID, testCase.lockType)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			for i, expectVideo := range testCase.expectVideos {
				assert.Equal(t, expectVideo.GetID(), videos[i].GetID())
				assert.Equal(t, expectVideo.GetType(), videos[i].GetType())
				assert.WithinDuration(t, expectVideo.GetCreatedAt(), videos[i].GetCreatedAt(), time.Second)
			}
		})
	}
}
