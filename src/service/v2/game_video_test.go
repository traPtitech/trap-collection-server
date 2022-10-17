package v2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
