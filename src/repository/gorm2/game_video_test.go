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
	"gorm.io/gorm"
)

func TestSetupVideoTypeTable(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		description      string
		beforeVideoTypes []string
		isErr            bool
		err              error
	}

	testCases := []test{
		{
			description:      "何も存在しない場合問題なし",
			beforeVideoTypes: []string{},
		},
		{
			description: "全て存在する場合問題なし",
			beforeVideoTypes: []string{
				gameVideoTypeMp4,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Delete(&GameVideoTypeTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete role type table: %+v\n", err)
				}
			}()

			if len(testCase.beforeVideoTypes) != 0 {
				videoTypes := make([]*GameVideoTypeTable, 0, len(testCase.beforeVideoTypes))
				for _, videoType := range testCase.beforeVideoTypes {
					videoTypes = append(videoTypes, &GameVideoTypeTable{
						Name: videoType,
					})
				}

				err := db.Create(videoTypes).Error
				if err != nil {
					t.Fatalf("failed to setup role type table: %+v\n", err)
				}
			}

			err := setupVideoTypeTable(db)

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

			var videoTypes []*GameVideoTypeTable
			err = db.
				Select("name").
				Find(&videoTypes).Error
			if err != nil {
				t.Fatalf("failed to get role type table: %+v\n", err)
			}

			videoTypeNames := make([]string, 0, len(videoTypes))
			for _, videoType := range videoTypes {
				videoTypeNames = append(videoTypeNames, videoType.Name)
			}

			assert.ElementsMatch(t, []string{
				gameVideoTypeMp4,
			}, videoTypeNames)
		})
	}
}

func TestSaveGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameVideoRepository, err := NewGameVideo(testDB)
	if err != nil {
		t.Fatalf("failed to create game video repository: %+v\n", err)
	}

	type test struct {
		description  string
		gameID       values.GameID
		video        *domain.GameVideo
		beforeVideos []GameVideoTable
		expectVideos []GameVideoTable
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

	var videoTypes []*GameVideoTypeTable
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

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			gameID:      gameID1,
			video: domain.NewGameVideo(
				videoID1,
				values.GameVideoTypeMp4,
			),
			beforeVideos: []GameVideoTable{},
			expectVideos: []GameVideoTable{
				{
					ID:          uuid.UUID(videoID1),
					GameID:      uuid.UUID(gameID1),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   time.Now(),
				},
			},
		},
		{
			description: "想定外の画像の種類なのでエラー",
			gameID:      gameID2,
			video: domain.NewGameVideo(
				videoID2,
				100,
			),
			beforeVideos: []GameVideoTable{},
			expectVideos: []GameVideoTable{},
			isErr:        true,
		},
		{
			description: "既に動画が存在しても問題なし",
			gameID:      gameID3,
			video: domain.NewGameVideo(
				videoID3,
				values.GameVideoTypeMp4,
			),
			beforeVideos: []GameVideoTable{
				{
					ID:          uuid.UUID(videoID4),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   time.Now().Add(-10 * time.Hour),
				},
			},
			expectVideos: []GameVideoTable{
				{
					ID:          uuid.UUID(videoID4),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   time.Now().Add(-10 * time.Hour),
				},
				{
					ID:          uuid.UUID(videoID3),
					GameID:      uuid.UUID(gameID3),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   time.Now(),
				},
			},
		},
		{
			description: "エラーの場合変更なし",
			gameID:      gameID4,
			video: domain.NewGameVideo(
				videoID5,
				100,
			),
			beforeVideos: []GameVideoTable{
				{
					ID:          uuid.UUID(videoID6),
					GameID:      uuid.UUID(gameID4),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   time.Now().Add(-10 * time.Hour),
				},
			},
			expectVideos: []GameVideoTable{
				{
					ID:          uuid.UUID(videoID6),
					GameID:      uuid.UUID(gameID4),
					VideoTypeID: videoTypeMap[gameVideoTypeMp4],
					CreatedAt:   time.Now().Add(-10 * time.Hour),
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&GameTable{
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

			var videos []GameVideoTable
			err = db.
				Session(&gorm.Session{}).
				Where("game_id = ?", uuid.UUID(testCase.gameID)).
				Find(&videos).Error
			if err != nil {
				t.Fatalf("failed to get role table: %+v\n", err)
			}

			assert.Len(t, videos, len(testCase.expectVideos))

			videoMap := make(map[uuid.UUID]GameVideoTable)
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
