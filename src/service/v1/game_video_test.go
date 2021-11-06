package v1

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	mockStorage "github.com/traPtitech/trap-collection-server/src/storage/mock"
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
			description: "動画が不正なのでエラー",
			gameID:      values.NewGameID(),
			isValidFile: false,
			isErr:       true,
			err:         service.ErrInvalidFormat,
		},
		{
			description:                    "repository.SaveGameVideoがエラーなのでエラー",
			gameID:                         values.NewGameID(),
			isValidFile:                    true,
			videoType:                      values.GameVideoTypeMp4,
			executeRepositorySaveGameVideo: true,
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
				videoBuf := bytes.NewBuffer(nil)

				switch testCase.videoType {
				case values.GameVideoTypeMp4:
					err := func() error {
						f, err := os.Open("../../../testdata/1.mp4")
						if err != nil {
							return fmt.Errorf("failed to open file: %w", err)
						}
						defer f.Close()

						_, err = io.Copy(videoBuf, f)
						if err != nil {
							return fmt.Errorf("failed to copy file: %w", err)
						}

						return nil
					}()
					if err != nil {
						t.Fatalf("failed to encode image: %s", err)
					}
				default:
					t.Fatalf("invalid video type: %v\n", testCase.videoType)
				}

				file = videoBuf
				expectBytes = videoBuf.Bytes()
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
					SaveGameVideo(ctx, testCase.gameID, gomock.Any()).
					Return(testCase.RepositorySaveGameVideoErr)
			}

			if testCase.executeStorageSaveGameVideo {
				mockGameVideoStorage.
					EXPECT().
					SaveGameVideo(ctx, gomock.Any()).
					Return(testCase.StorageSaveGameVideoErr)
			}

			err := gameVideoService.SaveGameVideo(ctx, file, testCase.gameID)

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

			assert.Equal(t, expectBytes, buf.Bytes())
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
		isValidFile               bool
		videoType                 values.GameVideoType
		executeGetLatestGameVideo bool
		video                     *domain.GameVideo
		GetLatestGameVideoErr     error
		executeGetGameVideo       bool
		GetGameVideoErr           error
		isErr                     bool
		err                       error
	}

	testCases := []test{
		{
			description:               "特に問題ないのでエラーなし",
			gameID:                    values.NewGameID(),
			isValidFile:               true,
			videoType:                 values.GameVideoTypeMp4,
			executeGetLatestGameVideo: true,
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
			),
			executeGetGameVideo: true,
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
			// 実際には発生しないが、念のため確認
			description:               "動画が不正でもエラーなし",
			gameID:                    values.NewGameID(),
			isValidFile:               false,
			executeGetLatestGameVideo: true,
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
			),
			executeGetGameVideo: true,
		},
		{
			description:               "GetLatestGameVideoがErrRecordNotFoundなのでErrNoGameImage",
			gameID:                    values.NewGameID(),
			isValidFile:               true,
			videoType:                 values.GameVideoTypeMp4,
			executeGetLatestGameVideo: true,
			GetLatestGameVideoErr:     repository.ErrRecordNotFound,
			isErr:                     true,
			err:                       service.ErrNoGameImage,
		},
		{
			description:               "GetLatestGameVideoがエラーなのでエラー",
			gameID:                    values.NewGameID(),
			isValidFile:               true,
			videoType:                 values.GameVideoTypeMp4,
			executeGetLatestGameVideo: true,
			GetLatestGameVideoErr:     errors.New("error"),
			isErr:                     true,
		},
		{
			description:               "GetGameVideoがエラーなのでエラー",
			gameID:                    values.NewGameID(),
			isValidFile:               true,
			videoType:                 values.GameVideoTypeMp4,
			executeGetLatestGameVideo: true,
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
			),
			executeGetGameVideo: true,
			GetGameVideoErr:     errors.New("error"),
			isErr:               true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var file *bytes.Buffer
			if testCase.isValidFile {
				videoBuf := bytes.NewBuffer(nil)

				switch testCase.videoType {
				case values.GameVideoTypeMp4:
					err := func() error {
						f, err := os.Open("../../../testdata/1.mp4")
						if err != nil {
							return fmt.Errorf("failed to open file: %w", err)
						}
						defer f.Close()

						_, err = io.Copy(videoBuf, f)
						if err != nil {
							return fmt.Errorf("failed to copy file: %w", err)
						}

						return nil
					}()
					if err != nil {
						t.Fatalf("failed to encode image: %s", err)
					}
				default:
					t.Fatalf("invalid video type: %v\n", testCase.videoType)
				}

				file = videoBuf
			} else {
				file = bytes.NewBufferString("invalid file")
			}
			expectBytes := file.Bytes()

			mockGameVideoStorage := mockStorage.NewGameVideo(ctrl, file)

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
					GetLatestGameVideo(ctx, testCase.gameID, repository.LockTypeRecord).
					Return(testCase.video, testCase.GetLatestGameVideoErr)
			}

			if testCase.executeGetGameVideo {
				mockGameVideoStorage.
					EXPECT().
					GetGameVideo(ctx, testCase.video).
					Return(testCase.GetGameVideoErr)
			}

			buf := bytes.NewBuffer(nil)

			err := gameVideoService.GetGameVideo(ctx, buf, testCase.gameID)

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

			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}
