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
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
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
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectVideos: []migrate.GameVideoTable2{
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
			beforeVideos: []migrate.GameVideoTable2{
				{
					ID:          uuid.UUID(videoID6),
					GameID:      uuid.UUID(gameID4),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectVideos: []migrate.GameVideoTable2{
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
			err := db.Create(&migrate.GameTable2{
				ID:          uuid.UUID(testCase.gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   time.Now(),
				GameVideo2s: testCase.beforeVideos,
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
