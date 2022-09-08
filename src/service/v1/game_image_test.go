package v1

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

func TestSaveGameImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameImageRepository := mockRepository.NewMockGameImage(ctrl)

	type test struct {
		description                    string
		gameID                         values.GameID
		isValidFile                    bool
		imageType                      values.GameImageType
		GetGameErr                     error
		executeRepositorySaveGameImage bool
		RepositorySaveGameImageErr     error
		executeStorageSaveGameImage    bool
		StorageSaveGameImageErr        error
		isErr                          bool
		err                            error
	}

	testCases := []test{
		{
			description:                    "特に問題ないのでエラーなし",
			gameID:                         values.NewGameID(),
			isValidFile:                    true,
			imageType:                      values.GameImageTypeJpeg,
			executeRepositorySaveGameImage: true,
			executeStorageSaveGameImage:    true,
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			isValidFile: true,
			imageType:   values.GameImageTypeJpeg,
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			isValidFile: true,
			imageType:   values.GameImageTypeJpeg,
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:                    "画像がpngでもエラーなし",
			gameID:                         values.NewGameID(),
			isValidFile:                    true,
			imageType:                      values.GameImageTypePng,
			executeRepositorySaveGameImage: true,
			executeStorageSaveGameImage:    true,
		},
		{
			description:                    "画像がgifでもエラーなし",
			gameID:                         values.NewGameID(),
			isValidFile:                    true,
			imageType:                      values.GameImageTypeGif,
			executeRepositorySaveGameImage: true,
			executeStorageSaveGameImage:    true,
		},
		{
			description:                 "画像が不正なのでエラー",
			gameID:                      values.NewGameID(),
			executeStorageSaveGameImage: true,
			isValidFile:                 false,
			isErr:                       true,
			err:                         service.ErrInvalidFormat,
		},
		{
			description:                    "repository.SaveGameImageがエラーなのでエラー",
			gameID:                         values.NewGameID(),
			isValidFile:                    true,
			imageType:                      values.GameImageTypeJpeg,
			executeRepositorySaveGameImage: true,
			executeStorageSaveGameImage:    true,
			RepositorySaveGameImageErr:     errors.New("error"),
			isErr:                          true,
		},
		{
			description:                    "storage.SaveGameImageがエラーなのでエラー",
			gameID:                         values.NewGameID(),
			isValidFile:                    true,
			imageType:                      values.GameImageTypeJpeg,
			executeRepositorySaveGameImage: true,
			executeStorageSaveGameImage:    true,
			StorageSaveGameImageErr:        errors.New("error"),
			isErr:                          true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)

			mockGameImageStorage := mockStorage.NewGameImage(ctrl, buf)

			gameImageService := NewGameImage(
				mockDB,
				mockGameRepository,
				mockGameImageRepository,
				mockGameImageStorage,
			)

			var file io.Reader
			var expectBytes []byte
			if testCase.isValidFile {
				imgBuf := bytes.NewBuffer(nil)

				err := func() error {
					var path string
					switch testCase.imageType {
					case values.GameImageTypeJpeg:
						path = "1.jpg"
					case values.GameImageTypePng:
						path = "1.png"
					case values.GameImageTypeGif:
						path = "1.gif"
					default:
						t.Fatalf("invalid image type: %v\n", testCase.imageType)
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
					t.Fatalf("failed to encode image: %s", err)
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

			if testCase.executeRepositorySaveGameImage {
				mockGameImageRepository.
					EXPECT().
					SaveGameImage(gomock.Any(), testCase.gameID, gomock.Any()).
					Return(testCase.RepositorySaveGameImageErr)
			}

			if testCase.executeStorageSaveGameImage {
				mockGameImageStorage.
					EXPECT().
					SaveGameImage(gomock.Any(), gomock.Any()).
					Return(testCase.StorageSaveGameImageErr)
			}

			err := gameImageService.SaveGameImage(ctx, file, testCase.gameID)

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

func TestGetGameImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameImageRepository := mockRepository.NewMockGameImage(ctrl)

	type test struct {
		description               string
		gameID                    values.GameID
		GetGameErr                error
		executeGetLatestGameImage bool
		image                     *domain.GameImage
		GetLatestGameImageErr     error
		executeGetTempURL         bool
		imageURL                  values.GameImageTmpURL
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
			executeGetLatestGameImage: true,
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
				time.Now(),
			),
			executeGetTempURL: true,
			imageURL:          values.NewGameImageTmpURL(urlLink),
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
			description:               "画像がpngでもエラーなし",
			gameID:                    values.NewGameID(),
			executeGetLatestGameImage: true,
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
				time.Now(),
			),
			executeGetTempURL: true,
		},
		{
			description:               "画像がgifでもエラーなし",
			gameID:                    values.NewGameID(),
			executeGetLatestGameImage: true,
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
				time.Now(),
			),
			executeGetTempURL: true,
		},
		{
			description:               "GetLatestGameImageがErrRecordNotFoundなのでErrNoGameImage",
			gameID:                    values.NewGameID(),
			executeGetLatestGameImage: true,
			GetLatestGameImageErr:     repository.ErrRecordNotFound,
			isErr:                     true,
			err:                       service.ErrNoGameImage,
		},
		{
			description:               "GetLatestGameImageがエラーなのでエラー",
			gameID:                    values.NewGameID(),
			executeGetLatestGameImage: true,
			GetLatestGameImageErr:     errors.New("error"),
			isErr:                     true,
		},
		{
			description:               "GetTempURLがエラーでもエラーなし",
			gameID:                    values.NewGameID(),
			executeGetLatestGameImage: true,
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
				time.Now(),
			),
			executeGetTempURL: true,
			imageURL:          values.NewGameImageTmpURL(urlLink),
			GetTempURLErr:     errors.New("error"),
			isErr:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameImageStorage := mockStorage.NewGameImage(ctrl, nil)

			gameImageService := NewGameImage(
				mockDB,
				mockGameRepository,
				mockGameImageRepository,
				mockGameImageStorage,
			)

			mockGameRepository.
				EXPECT().
				GetGame(ctx, testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.GetGameErr)

			if testCase.executeGetLatestGameImage {
				mockGameImageRepository.
					EXPECT().
					GetLatestGameImage(ctx, testCase.gameID, repository.LockTypeNone).
					Return(testCase.image, testCase.GetLatestGameImageErr)
			}

			if testCase.executeGetTempURL {
				mockGameImageStorage.
					EXPECT().
					GetTempURL(ctx, testCase.image, time.Minute).
					Return(testCase.imageURL, testCase.GetTempURLErr)
			}

			tmpURL, err := gameImageService.GetGameImage(ctx, testCase.gameID)

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

			assert.Equal(t, testCase.imageURL, tmpURL)
		})
	}
}
