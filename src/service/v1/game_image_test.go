package v1

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
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
			description: "画像が不正なのでエラー",
			gameID:      values.NewGameID(),
			isValidFile: false,
			isErr:       true,
			err:         service.ErrInvalidFormat,
		},
		{
			description:                    "repository.SaveGameImageがエラーなのでエラー",
			gameID:                         values.NewGameID(),
			isValidFile:                    true,
			imageType:                      values.GameImageTypeJpeg,
			executeRepositorySaveGameImage: true,
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
				img := image.NewRGBA(image.Rect(0, 0, 3000, 3000))
				imgBuf := bytes.NewBuffer(nil)

				switch testCase.imageType {
				case values.GameImageTypeJpeg:
					err := jpeg.Encode(imgBuf, img, nil)
					if err != nil {
						t.Fatalf("failed to encode image: %v\n", err)
					}
				case values.GameImageTypePng:
					err := png.Encode(imgBuf, img)
					if err != nil {
						t.Fatalf("failed to encode image: %v\n", err)
					}
				case values.GameImageTypeGif:
					err := gif.Encode(imgBuf, img, nil)
					if err != nil {
						t.Fatalf("failed to encode image: %v\n", err)
					}
				default:
					t.Fatalf("invalid image type: %v\n", testCase.imageType)
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
					SaveGameImage(ctx, testCase.gameID, gomock.Any()).
					Return(testCase.RepositorySaveGameImageErr)
			}

			if testCase.executeStorageSaveGameImage {
				mockGameImageStorage.
					EXPECT().
					SaveGameImage(ctx, gomock.Any()).
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
		description                   string
		gameID                        values.GameID
		GetGameErr                    error
		isValidFile                   bool
		imageType                     values.GameImageType
		executeRepositoryGetGameImage bool
		image                         *domain.GameImage
		RepositoryGetGameImageErr     error
		executeStorageGetGameImage    bool
		StorageGetGameImageErr        error
		isErr                         bool
		err                           error
	}

	testCases := []test{
		{
			description:                   "特に問題ないのでエラーなし",
			gameID:                        values.NewGameID(),
			isValidFile:                   true,
			imageType:                     values.GameImageTypeJpeg,
			executeRepositoryGetGameImage: true,
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
			),
			executeStorageGetGameImage: true,
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
			description:                   "画像がpngでもエラーなし",
			gameID:                        values.NewGameID(),
			isValidFile:                   true,
			imageType:                     values.GameImageTypePng,
			executeRepositoryGetGameImage: true,
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
			),
			executeStorageGetGameImage: true,
		},
		{
			description:                   "画像がgifでもエラーなし",
			gameID:                        values.NewGameID(),
			isValidFile:                   true,
			imageType:                     values.GameImageTypeGif,
			executeRepositoryGetGameImage: true,
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
			),
			executeStorageGetGameImage: true,
		},
		{
			// 実際には発生しないが、念のため確認
			description:                   "画像が不正でもエラーなし",
			gameID:                        values.NewGameID(),
			isValidFile:                   false,
			executeRepositoryGetGameImage: true,
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
			),
			executeStorageGetGameImage: true,
		},
		{
			description:                   "repository.GetGameImageがErrRecordNotFoundなのでErrNoGameImage",
			gameID:                        values.NewGameID(),
			isValidFile:                   true,
			imageType:                     values.GameImageTypeJpeg,
			executeRepositoryGetGameImage: true,
			RepositoryGetGameImageErr:     repository.ErrRecordNotFound,
			isErr:                         true,
			err:                           service.ErrNoGameImage,
		},
		{
			description:                   "repository.GetGameImageがエラーなのでエラー",
			gameID:                        values.NewGameID(),
			isValidFile:                   true,
			imageType:                     values.GameImageTypeJpeg,
			executeRepositoryGetGameImage: true,
			RepositoryGetGameImageErr:     errors.New("error"),
			isErr:                         true,
		},
		{
			description:                   "storage.GetGameImageがエラーなのでエラー",
			gameID:                        values.NewGameID(),
			isValidFile:                   true,
			imageType:                     values.GameImageTypeJpeg,
			executeRepositoryGetGameImage: true,
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
			),
			executeStorageGetGameImage: true,
			StorageGetGameImageErr:     errors.New("error"),
			isErr:                      true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var file *bytes.Buffer
			if testCase.isValidFile {
				img := image.NewRGBA(image.Rect(0, 0, 100, 100))
				imgBuf := bytes.NewBuffer(nil)

				switch testCase.imageType {
				case values.GameImageTypeJpeg:
					err := jpeg.Encode(imgBuf, img, nil)
					if err != nil {
						t.Fatalf("failed to encode image: %v\n", err)
					}
				case values.GameImageTypePng:
					err := png.Encode(imgBuf, img)
					if err != nil {
						t.Fatalf("failed to encode image: %v\n", err)
					}
				case values.GameImageTypeGif:
					err := gif.Encode(imgBuf, img, nil)
					if err != nil {
						t.Fatalf("failed to encode image: %v\n", err)
					}
				default:
					t.Fatalf("invalid image type: %v\n", testCase.imageType)
				}

				file = imgBuf
			} else {
				file = bytes.NewBufferString("invalid file")
			}
			expectBytes := file.Bytes()

			mockGameImageStorage := mockStorage.NewGameImage(ctrl, file)

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

			if testCase.executeRepositoryGetGameImage {
				mockGameImageRepository.
					EXPECT().
					GetGameImage(ctx, testCase.gameID, repository.LockTypeRecord).
					Return(testCase.image, testCase.RepositoryGetGameImageErr)
			}

			if testCase.executeStorageGetGameImage {
				mockGameImageStorage.
					EXPECT().
					GetGameImage(ctx, testCase.image).
					Return(testCase.StorageGetGameImageErr)
			}

			buf := bytes.NewBuffer(nil)

			err := gameImageService.GetGameImage(ctx, buf, testCase.gameID)

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
