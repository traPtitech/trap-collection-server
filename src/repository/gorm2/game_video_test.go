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

func TestSaveGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameVideoRepository := NewGameVideo(testDB)

	type test struct {
		description  string
		gameID       values.GameID
		video        *domain.GameVideo
		beforeVideos []migrate.GameVideoTable
		expectVideos []migrate.GameVideoTable
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
			beforeVideos: []migrate.GameVideoTable{},
			expectVideos: []migrate.GameVideoTable{
				{
					ID:          uuid.UUID(videoID1),
					GameID:      uuid.UUID(gameID1),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now,
				},
			},
		},
		{
			description: "想定外の画像の種類なのでエラー",
			gameID:      gameID2,
			video: domain.NewGameVideo(
				videoID2,
				100,
				now,
			),
			beforeVideos: []migrate.GameVideoTable{},
			expectVideos: []migrate.GameVideoTable{},
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
			beforeVideos: []migrate.GameVideoTable{
				{
					ID:          uuid.UUID(videoID4),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectVideos: []migrate.GameVideoTable{
				{
					ID:          uuid.UUID(videoID4),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
				{
					ID:          uuid.UUID(videoID3),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
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
			beforeVideos: []migrate.GameVideoTable{
				{
					ID:          uuid.UUID(videoID6),
					GameID:      uuid.UUID(gameID4),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectVideos: []migrate.GameVideoTable{
				{
					ID:          uuid.UUID(videoID6),
					GameID:      uuid.UUID(gameID4),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&migrate.GameTable{
				ID:          uuid.UUID(testCase.gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   time.Now(),
				GameVideos:  testCase.beforeVideos,
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

			var videos []migrate.GameVideoTable
			err = db.
				Session(&gorm.Session{}).
				Where("game_id = ?", uuid.UUID(testCase.gameID)).
				Find(&videos).Error
			if err != nil {
				t.Fatalf("failed to get role table: %+v\n", err)
			}

			assert.Len(t, videos, len(testCase.expectVideos))

			videoMap := make(map[uuid.UUID]migrate.GameVideoTable)
			for _, video := range videos {
				videoMap[video.ID] = video
			}

			for _, expectVideo := range testCase.expectVideos {
				actualVideo, ok := videoMap[expectVideo.ID]
				if !ok {
					t.Errorf("not found image: %+v", expectVideo)
				}

				assert.Equal(t, expectVideo.GameID, actualVideo.GameID)
				assert.Equal(t, expectVideo.VideoTypeID, actualVideo.VideoTypeID)
				assert.WithinDuration(t, expectVideo.CreatedAt, actualVideo.CreatedAt, 2*time.Second)
			}
		})
	}
}

func TestGetLatestGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameVideoRepository := NewGameVideo(testDB)

	type test struct {
		description string
		gameID      values.GameID
		lockType    repository.LockType
		videos      []migrate.GameVideoTable
		expectVideo *domain.GameVideo
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()

	videoID1 := values.NewGameVideoID()
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

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			lockType:    repository.LockTypeNone,
			videos: []migrate.GameVideoTable{
				{
					ID:          uuid.UUID(videoID1),
					GameID:      uuid.UUID(gameID1),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now,
				},
			},
			expectVideo: domain.NewGameVideo(
				videoID1,
				values.GameVideoTypeMp4,
				now,
			),
		},
		{
			description: "画像がないのでエラー",
			gameID:      gameID4,
			lockType:    repository.LockTypeNone,
			videos:      []migrate.GameVideoTable{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			description: "複数でも正しい画像が返る",
			gameID:      gameID5,
			lockType:    repository.LockTypeNone,
			videos: []migrate.GameVideoTable{
				{
					ID:          uuid.UUID(videoID4),
					GameID:      uuid.UUID(gameID5),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now.Add(-24 * time.Hour),
				},
				{
					ID:          uuid.UUID(videoID5),
					GameID:      uuid.UUID(gameID5),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now,
				},
			},
			expectVideo: domain.NewGameVideo(
				videoID5,
				values.GameVideoTypeMp4,
				now,
			),
		},
		{
			description: "行ロックをとっても問題なし",
			gameID:      gameID6,
			lockType:    repository.LockTypeRecord,
			videos: []migrate.GameVideoTable{
				{
					ID:          uuid.UUID(videoID6),
					GameID:      uuid.UUID(gameID6),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now,
				},
			},
			expectVideo: domain.NewGameVideo(
				videoID6,
				values.GameVideoTypeMp4,
				now,
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&migrate.GameTable{
				ID:          uuid.UUID(testCase.gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   time.Now(),
				GameVideos:  testCase.videos,
			}).Error
			if err != nil {
				t.Fatalf("failed to create game table: %+v\n", err)
			}

			video, err := gameVideoRepository.GetLatestGameVideo(ctx, testCase.gameID, testCase.lockType)

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

			assert.Equal(t, testCase.expectVideo.GetID(), video.GetID())
			assert.Equal(t, testCase.expectVideo.GetType(), video.GetType())
			assert.WithinDuration(t, testCase.expectVideo.GetCreatedAt(), video.GetCreatedAt(), time.Second)
		})
	}
}
