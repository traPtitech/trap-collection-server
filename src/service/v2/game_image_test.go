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

func TestSaveGameImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameImageRepository := mockRepository.NewMockGameImageV2(ctrl)

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
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}

func TestGetGameImages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameImageRepository := mockRepository.NewMockGameImageV2(ctrl)
	mockGameImageStorage := mockStorage.NewGameImage(ctrl, nil)

	gameImageService := NewGameImage(
		mockDB,
		mockGameRepository,
		mockGameImageRepository,
		mockGameImageStorage,
	)

	type test struct {
		description          string
		gameID               values.GameID
		getGameErr           error
		executeGetGameImages bool
		getGameImagesErr     error
		isErr                bool
		gameImages           []*domain.GameImage
		err                  error
	}

	now := time.Now()
	testCases := []test{
		{
			description:          "特に問題ないのでエラーなし",
			gameID:               values.NewGameID(),
			executeGetGameImages: true,
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
			description:          "画像がpngでもエラーなし",
			gameID:               values.NewGameID(),
			executeGetGameImages: true,
			gameImages: []*domain.GameImage{
				domain.NewGameImage(
					values.NewGameImageID(),
					values.GameImageTypePng,
					now,
				),
			},
		},
		{
			description:          "画像がgifでもエラーなし",
			gameID:               values.NewGameID(),
			executeGetGameImages: true,
			gameImages: []*domain.GameImage{
				domain.NewGameImage(
					values.NewGameImageID(),
					values.GameImageTypeGif,
					now,
				),
			},
		},
		{
			description:          "画像がなくてもエラーなし",
			gameID:               values.NewGameID(),
			executeGetGameImages: true,
			gameImages:           []*domain.GameImage{},
		},
		{
			description:          "画像が複数でもエラーなし",
			gameID:               values.NewGameID(),
			executeGetGameImages: true,
			gameImages: []*domain.GameImage{
				domain.NewGameImage(
					values.NewGameImageID(),
					values.GameImageTypePng,
					now,
				),
				domain.NewGameImage(
					values.NewGameImageID(),
					values.GameImageTypeGif,
					now.Add(-time.Second),
				),
			},
		},
		{
			description:          "GetGameImagesがエラーなのでエラー",
			gameID:               values.NewGameID(),
			executeGetGameImages: true,
			getGameImagesErr:     errors.New("error"),
			isErr:                true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.getGameErr)

			if testCase.executeGetGameImages {
				mockGameImageRepository.
					EXPECT().
					GetGameImages(gomock.Any(), testCase.gameID, repository.LockTypeNone).
					Return(testCase.gameImages, testCase.getGameImagesErr)
			}

			gameImages, err := gameImageService.GetGameImages(ctx, testCase.gameID)

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

			for i, gameImage := range gameImages {
				assert.Equal(t, testCase.gameImages[i].GetID(), gameImage.GetID())
				assert.Equal(t, testCase.gameImages[i].GetType(), gameImage.GetType())
				assert.Equal(t, testCase.gameImages[i].GetCreatedAt(), gameImage.GetCreatedAt())
			}
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
	mockGameImageRepository := mockRepository.NewMockGameImageV2(ctrl)
	mockGameImageStorage := mockStorage.NewGameImage(ctrl, nil)

	gameImageService := NewGameImage(
		mockDB,
		mockGameRepository,
		mockGameImageRepository,
		mockGameImageStorage,
	)

	type test struct {
		description         string
		gameID              values.GameID
		gameImageID         values.GameImageID
		getGameErr          error
		executeGetGameImage bool
		image               *repository.GameImageInfo
		getGameImageErr     error
		executeGetTempURL   bool
		imageURL            values.GameImageTmpURL
		getTempURLErr       error
		isErr               bool
		err                 error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()

	testCases := []test{
		{
			description:         "特に問題ないのでエラーなし",
			gameID:              gameID1,
			executeGetGameImage: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					values.NewGameImageID(),
					values.GameImageTypeJpeg,
					time.Now(),
				),
				GameID: gameID1,
			},
			executeGetTempURL: true,
			imageURL:          values.NewGameImageTmpURL(urlLink),
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
			description:         "画像がpngでもエラーなし",
			gameID:              gameID2,
			executeGetGameImage: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					values.NewGameImageID(),
					values.GameImageTypePng,
					time.Now(),
				),
				GameID: gameID2,
			},
			executeGetTempURL: true,
		},
		{
			description:         "画像がgifでもエラーなし",
			gameID:              gameID3,
			executeGetGameImage: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					values.NewGameImageID(),
					values.GameImageTypeGif,
					time.Now(),
				),
				GameID: gameID3,
			},
			executeGetTempURL: true,
		},
		{
			description:         "GetGameImageがErrRecordNotFoundなのでErrInvalidGameImageID",
			gameID:              values.NewGameID(),
			executeGetGameImage: true,
			getGameImageErr:     repository.ErrRecordNotFound,
			isErr:               true,
			err:                 service.ErrInvalidGameImageID,
		},
		{
			description:         "ゲーム画像に紐づくゲームIDが違うのでErrInvalidGameImageID",
			gameID:              values.NewGameID(),
			executeGetGameImage: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					values.NewGameImageID(),
					values.GameImageTypeGif,
					time.Now(),
				),
				GameID: values.NewGameID(),
			},
			isErr: true,
			err:   service.ErrInvalidGameImageID,
		},
		{
			description:         "GetGameImageがエラーなのでエラー",
			gameID:              values.NewGameID(),
			executeGetGameImage: true,
			getGameImageErr:     errors.New("error"),
			isErr:               true,
		},
		{
			description:         "GetTempURLがエラーなのでエラー",
			gameID:              gameID4,
			executeGetGameImage: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					values.NewGameImageID(),
					values.GameImageTypeJpeg,
					time.Now(),
				),
				GameID: gameID4,
			},
			executeGetTempURL: true,
			imageURL:          values.NewGameImageTmpURL(urlLink),
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

			if testCase.executeGetGameImage {
				mockGameImageRepository.
					EXPECT().
					GetGameImage(ctx, testCase.gameImageID, repository.LockTypeRecord).
					Return(testCase.image, testCase.getGameImageErr)
			}

			if testCase.executeGetTempURL {
				mockGameImageStorage.
					EXPECT().
					GetTempURL(ctx, testCase.image.GameImage, time.Minute).
					Return(testCase.imageURL, testCase.getTempURLErr)
			}

			tmpURL, err := gameImageService.GetGameImage(ctx, testCase.gameID, testCase.gameImageID)

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
