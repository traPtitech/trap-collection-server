package v2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	mockStorage "github.com/traPtitech/trap-collection-server/src/storage/mock"
	"github.com/traPtitech/trap-collection-server/testdata"
)

func TestSaveGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameVideoRepository := mockRepository.NewMockGameVideoV2(ctrl)

	type test struct {
		description                    string
		gameID                         values.GameID
		isValidFile                    bool
		videoType                      values.GameVideoType
		GetGameErr                     error
		executeRepositorySaveGameVideo bool
		RepositorySaveGameVideoErr     error
		executeStorageSaveGameVideo    bool
		StorageSaveGameVideoErr        error
		isErr                          bool
		err                            error
	}

	testCases := []test{
		{
			description:                    "特に問題ないのでエラーなし",
			gameID:                         values.NewGameID(),
			isValidFile:                    true,
			videoType:                      values.GameVideoTypeMp4,
			executeRepositorySaveGameVideo: true,
			executeStorageSaveGameVideo:    true,
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			isValidFile: true,
			videoType:   values.GameVideoTypeMp4,
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			isValidFile: true,
			videoType:   values.GameVideoTypeMp4,
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:                 "動画が不正なのでエラー",
			gameID:                      values.NewGameID(),
			executeStorageSaveGameVideo: true,
			isValidFile:                 false,
			isErr:                       true,
			err:                         service.ErrInvalidFormat,
		},
		{
			description:                    "repository.SaveGameVideoがエラーなのでエラー",
			gameID:                         values.NewGameID(),
			isValidFile:                    true,
			videoType:                      values.GameVideoTypeMp4,
			executeRepositorySaveGameVideo: true,
			executeStorageSaveGameVideo:    true,
			RepositorySaveGameVideoErr:     errors.New("error"),
			isErr:                          true,
		},
		{
			description:                    "storage.SaveGameVideoがエラーなのでエラー",
			gameID:                         values.NewGameID(),
			isValidFile:                    true,
			videoType:                      values.GameVideoTypeMp4,
			executeRepositorySaveGameVideo: true,
			executeStorageSaveGameVideo:    true,
			StorageSaveGameVideoErr:        errors.New("error"),
			isErr:                          true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)

			mockGameVideoStorage := mockStorage.NewGameVideo(ctrl, buf)

			gameVideoService := NewGameVideo(
				mockDB,
				mockGameRepository,
				mockGameVideoRepository,
				mockGameVideoStorage,
			)

			var file io.Reader
			var expectBytes []byte
			if testCase.isValidFile {
				imgBuf := bytes.NewBuffer(nil)

				err := func() error {
					var path string
					if testCase.videoType == values.GameVideoTypeMp4 {
						path = "1.mp4"
					} else {
						t.Fatalf("invalid video type: %v\n", testCase.videoType)
					}

					f, err := testdata.FS.Open(path)
					if err != nil {
						return fmt.Errorf("failed to open file: %w", err)
					}
					defer f.Close()

					_, err = io.Copy(imgBuf, f)
					if err != nil {
						return fmt.Errorf("failed to copy file: %w", err)
					}

					return nil
				}()
				if err != nil {
					t.Fatalf("failed to encode video: %s", err)
				}

				file = imgBuf
				expectBytes = imgBuf.Bytes()
			} else {
				file = strings.NewReader("invalid file")
			}

			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeRecord).
				Return(nil, testCase.GetGameErr)

			if testCase.executeRepositorySaveGameVideo {
				mockGameVideoRepository.
					EXPECT().
					SaveGameVideo(gomock.Any(), testCase.gameID, gomock.Any()).
					Return(testCase.RepositorySaveGameVideoErr)
			}

			if testCase.executeStorageSaveGameVideo {
				mockGameVideoStorage.
					EXPECT().
					SaveGameVideo(gomock.Any(), gomock.Any()).
					Return(testCase.StorageSaveGameVideoErr)
			}

			video, err := gameVideoService.SaveGameVideo(ctx, file, testCase.gameID)

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

			assert.Equal(t, testCase.videoType, video.GetType())
			assert.WithinDuration(t, time.Now(), video.GetCreatedAt(), time.Second)

			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}

func TestGetGameVideos(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameVideoRepository := mockRepository.NewMockGameVideoV2(ctrl)
	mockGameVideoStorage := mockStorage.NewGameVideo(ctrl, nil)

	gameVideoService := NewGameVideo(
		mockDB,
		mockGameRepository,
		mockGameVideoRepository,
		mockGameVideoStorage,
	)

	type test struct {
		description          string
		gameID               values.GameID
		getGameErr           error
		executeGetGameVideos bool
		getGameVideosErr     error
		isErr                bool
		gameVideos           []*domain.GameVideo
		err                  error
	}

	now := time.Now()
	testCases := []test{
		{
			description:          "特に問題ないのでエラーなし",
			gameID:               values.NewGameID(),
			executeGetGameVideos: true,
			gameVideos: []*domain.GameVideo{
				domain.NewGameVideo(
					values.NewGameVideoID(),
					values.GameVideoTypeMp4,
					now,
				),
			},
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			getGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			getGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:          "動画がなくてもエラーなし",
			gameID:               values.NewGameID(),
			executeGetGameVideos: true,
			gameVideos:           []*domain.GameVideo{},
		},
		{
			description:          "動画が複数でもエラーなし",
			gameID:               values.NewGameID(),
			executeGetGameVideos: true,
			gameVideos: []*domain.GameVideo{
				domain.NewGameVideo(
					values.NewGameVideoID(),
					values.GameVideoTypeMp4,
					now,
				),
				domain.NewGameVideo(
					values.NewGameVideoID(),
					values.GameVideoTypeMp4,
					now.Add(-time.Second),
				),
			},
		},
		{
			description:          "GetGameVideosがエラーなのでエラー",
			gameID:               values.NewGameID(),
			executeGetGameVideos: true,
			getGameVideosErr:     errors.New("error"),
			isErr:                true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.getGameErr)

			if testCase.executeGetGameVideos {
				mockGameVideoRepository.
					EXPECT().
					GetGameVideos(gomock.Any(), testCase.gameID, repository.LockTypeNone).
					Return(testCase.gameVideos, testCase.getGameVideosErr)
			}

			gameVideos, err := gameVideoService.GetGameVideos(ctx, testCase.gameID)

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

			for i, gameVideo := range gameVideos {
				assert.Equal(t, testCase.gameVideos[i].GetID(), gameVideo.GetID())
				assert.Equal(t, testCase.gameVideos[i].GetType(), gameVideo.GetType())
				assert.Equal(t, testCase.gameVideos[i].GetCreatedAt(), gameVideo.GetCreatedAt())
			}
		})
	}
}

func TestGetGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameVideoRepository := mockRepository.NewMockGameVideoV2(ctrl)
	mockGameVideoStorage := mockStorage.NewGameVideo(ctrl, nil)

	gameVideoService := NewGameVideo(
		mockDB,
		mockGameRepository,
		mockGameVideoRepository,
		mockGameVideoStorage,
	)

	type test struct {
		description         string
		gameID              values.GameID
		gameVideoID         values.GameVideoID
		getGameErr          error
		executeGetGameVideo bool
		video               *repository.GameVideoInfo
		getGameVideoErr     error
		executeGetTempURL   bool
		videoURL            values.GameVideoTmpURL
		getTempURLErr       error
		isErr               bool
		err                 error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode video: %v", err)
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	testCases := []test{
		{
			description:         "特に問題ないのでエラーなし",
			gameID:              gameID1,
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					values.NewGameVideoID(),
					values.GameVideoTypeMp4,
					time.Now(),
				),
				GameID: gameID1,
			},
			executeGetTempURL: true,
			videoURL:          values.NewGameVideoTmpURL(urlLink),
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			getGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			getGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:         "GetGameVideoがErrRecordNotFoundなのでErrInvalidGameVideoID",
			gameID:              values.NewGameID(),
			executeGetGameVideo: true,
			getGameVideoErr:     repository.ErrRecordNotFound,
			isErr:               true,
			err:                 service.ErrInvalidGameVideoID,
		},
		{
			description:         "ゲーム動画に紐づくゲームIDが違うのでErrInvalidGameVideoID",
			gameID:              values.NewGameID(),
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					values.NewGameVideoID(),
					values.GameVideoTypeMp4,
					time.Now(),
				),
				GameID: values.NewGameID(),
			},
			isErr: true,
			err:   service.ErrInvalidGameVideoID,
		},
		{
			description:         "GetGameVideoがエラーなのでエラー",
			gameID:              values.NewGameID(),
			executeGetGameVideo: true,
			getGameVideoErr:     errors.New("error"),
			isErr:               true,
		},
		{
			description:         "GetTempURLがエラーなのでエラー",
			gameID:              gameID2,
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					values.NewGameVideoID(),
					values.GameVideoTypeMp4,
					time.Now(),
				),
				GameID: gameID2,
			},
			executeGetTempURL: true,
			videoURL:          values.NewGameVideoTmpURL(urlLink),
			getTempURLErr:     errors.New("error"),
			isErr:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(ctx, testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.getGameErr)

			if testCase.executeGetGameVideo {
				mockGameVideoRepository.
					EXPECT().
					GetGameVideo(ctx, testCase.gameVideoID, repository.LockTypeRecord).
					Return(testCase.video, testCase.getGameVideoErr)
			}

			if testCase.executeGetTempURL {
				mockGameVideoStorage.
					EXPECT().
					GetTempURL(ctx, testCase.video.GameVideo, time.Minute).
					Return(testCase.videoURL, testCase.getTempURLErr)
			}

			tmpURL, err := gameVideoService.GetGameVideo(ctx, testCase.gameID, testCase.gameVideoID)

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

			assert.Equal(t, testCase.videoURL, tmpURL)
		})
	}
}

func TestGetGameVideoMeta(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameVideoRepository := mockRepository.NewMockGameVideoV2(ctrl)
	mockGameVideoStorage := mockStorage.NewGameVideo(ctrl, nil)

	gameVideoService := NewGameVideo(
		mockDB,
		mockGameRepository,
		mockGameVideoRepository,
		mockGameVideoStorage,
	)

	type test struct {
		description         string
		gameID              values.GameID
		gameVideoID         values.GameVideoID
		getGameErr          error
		executeGetGameVideo bool
		video               *repository.GameVideoInfo
		getGameVideoErr     error
		isErr               bool
		err                 error
	}

	gameID1 := values.NewGameID()

	testCases := []test{
		{
			description:         "特に問題ないのでエラーなし",
			gameID:              gameID1,
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					values.NewGameVideoID(),
					values.GameVideoTypeMp4,
					time.Now(),
				),
				GameID: gameID1,
			},
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			getGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			getGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:         "GetGameVideoがErrRecordNotFoundなのでErrInvalidGameVideoID",
			gameID:              values.NewGameID(),
			executeGetGameVideo: true,
			getGameVideoErr:     repository.ErrRecordNotFound,
			isErr:               true,
			err:                 service.ErrInvalidGameVideoID,
		},
		{
			description:         "ゲーム動画に紐づくゲームIDが違うのでErrInvalidGameVideoID",
			gameID:              values.NewGameID(),
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					values.NewGameVideoID(),
					values.GameVideoTypeMp4,
					time.Now(),
				),
				GameID: values.NewGameID(),
			},
			isErr: true,
			err:   service.ErrInvalidGameVideoID,
		},
		{
			description:         "GetGameVideoがエラーなのでエラー",
			gameID:              values.NewGameID(),
			executeGetGameVideo: true,
			getGameVideoErr:     errors.New("error"),
			isErr:               true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(ctx, testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.getGameErr)

			if testCase.executeGetGameVideo {
				mockGameVideoRepository.
					EXPECT().
					GetGameVideo(ctx, testCase.gameVideoID, repository.LockTypeNone).
					Return(testCase.video, testCase.getGameVideoErr)
			}

			gameVideo, err := gameVideoService.GetGameVideoMeta(ctx, testCase.gameID, testCase.gameVideoID)

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

			assert.Equal(t, testCase.video.GameVideo.GetID(), gameVideo.GetID())
			assert.Equal(t, testCase.video.GameVideo.GetType(), gameVideo.GetType())
			assert.Equal(t, testCase.video.GameVideo.GetCreatedAt(), gameVideo.GetCreatedAt())
		})
	}
}
