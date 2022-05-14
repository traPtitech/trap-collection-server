package v1

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	pkgio "github.com/traPtitech/trap-collection-server/pkg/io"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	mockStorage "github.com/traPtitech/trap-collection-server/src/storage/mock"
	"golang.org/x/sync/errgroup"
)

func TestSaveGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVideoRepository := mockRepository.NewMockGameVideo(ctrl)

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
			isValidFile:                 false,
			executeStorageSaveGameVideo: true,
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
			pr, pw := io.Pipe()
			defer pr.Close()

			mockGameVideoStorage := mockStorage.NewGameVideo(ctrl, pw)

			gameVideoService := NewGameVideo(
				mockDB,
				mockGameRepository,
				mockGameVideoRepository,
				mockGameVideoStorage,
			)

			var file io.Reader
			var expectFile io.Reader
			if testCase.isValidFile {
				switch testCase.videoType {
				case values.GameVideoTypeMp4:
					f, err := os.Open("../../../testdata/1.mp4")
					if err != nil {
						t.Fatalf("failed to open file: %s", err)
						return
					}
					defer f.Close()

					file = f

					f, err = os.Open("../../../testdata/1.mp4")
					if err != nil {
						t.Fatalf("failed to open file: %s", err)
						return
					}
					defer f.Close()

					expectFile = f
				default:
					t.Fatalf("invalid video type: %v\n", testCase.videoType)
				}
			} else {
				file = strings.NewReader("invalid file")
				expectFile = strings.NewReader("invalid file")
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

			eg := errgroup.Group{}
			eg.Go(func() error {
				defer pw.Close()

				err := gameVideoService.SaveGameVideo(ctx, file, testCase.gameID)
				if err != nil {
					return err
				}

				return nil
			})

			isEqual, err := pkgio.ReaderEqual(expectFile, pr)
			if err != nil {
				t.Fatalf("failed to compare file: %s", err)
				return
			}

			err = eg.Wait()
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

			assert.True(t, isEqual)
		})
	}
}

func TestGetGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVideoRepository := mockRepository.NewMockGameVideo(ctrl)

	type test struct {
		description               string
		gameID                    values.GameID
		GetGameErr                error
		executeGetLatestGameVideo bool
		video                     *domain.GameVideo
		GetLatestGameVideoErr     error
		executeGetTempURL         bool
		videoURL                  values.GameVideoTmpURL
		GetTempURLErr             error
		isErr                     bool
		err                       error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	testCases := []test{
		{
			description:               "特に問題ないのでエラーなし",
			gameID:                    values.NewGameID(),
			executeGetLatestGameVideo: true,
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
				time.Now(),
			),
			executeGetTempURL: true,
			videoURL:          values.NewGameVideoTmpURL(urlLink),
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:               "GetLatestGameVideoがErrRecordNotFoundなのでErrNoGameImage",
			gameID:                    values.NewGameID(),
			executeGetLatestGameVideo: true,
			GetLatestGameVideoErr:     repository.ErrRecordNotFound,
			isErr:                     true,
			err:                       service.ErrNoGameVideo,
		},
		{
			description:               "GetLatestGameVideoがエラーなのでエラー",
			gameID:                    values.NewGameID(),
			executeGetLatestGameVideo: true,
			GetLatestGameVideoErr:     errors.New("error"),
			isErr:                     true,
		},
		{
			description:               "GetTempURLがエラーなのでエラー",
			gameID:                    values.NewGameID(),
			executeGetLatestGameVideo: true,
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
				time.Now(),
			),
			executeGetTempURL: true,
			videoURL:          values.NewGameVideoTmpURL(urlLink),
			GetTempURLErr:     errors.New("error"),
			isErr:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameVideoStorage := mockStorage.NewGameVideo(ctrl, nil)

			gameVideoService := NewGameVideo(
				mockDB,
				mockGameRepository,
				mockGameVideoRepository,
				mockGameVideoStorage,
			)

			mockGameRepository.
				EXPECT().
				GetGame(ctx, testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.GetGameErr)

			if testCase.executeGetLatestGameVideo {
				mockGameVideoRepository.
					EXPECT().
					GetLatestGameVideo(ctx, testCase.gameID, repository.LockTypeNone).
					Return(testCase.video, testCase.GetLatestGameVideoErr)
			}

			if testCase.executeGetTempURL {
				mockGameVideoStorage.
					EXPECT().
					GetTempURL(ctx, testCase.video, time.Minute).
					Return(testCase.videoURL, testCase.GetTempURLErr)
			}

			tmpURL, err := gameVideoService.GetGameVideo(ctx, testCase.gameID)

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
