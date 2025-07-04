package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestGetGameImages(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameImageService := mock.NewMockGameImageV2(ctrl)

	gameImage := NewGameImage(mockGameImageService)

	type test struct {
		description      string
		gameID           openapi.GameIDInPath
		images           []*domain.GameImage
		getGameImagesErr error
		resImages        []openapi.GameImage
		isErr            bool
		err              error
		statusCode       int
	}

	gameImageID1 := values.NewGameImageID()
	gameImageID2 := values.NewGameImageID()
	gameImageID3 := values.NewGameImageID()
	gameImageID4 := values.NewGameImageID()
	gameImageID5 := values.NewGameImageID()

	now := time.Now()
	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			images: []*domain.GameImage{
				domain.NewGameImage(
					gameImageID1,
					values.GameImageTypeJpeg,
					now,
				),
			},
			resImages: []openapi.GameImage{
				{
					Id:        uuid.UUID(gameImageID1),
					Mime:      openapi.Imagejpeg,
					CreatedAt: now,
				},
			},
		},
		{
			description: "pngでもエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			images: []*domain.GameImage{
				domain.NewGameImage(
					gameImageID2,
					values.GameImageTypePng,
					now,
				),
			},
			resImages: []openapi.GameImage{
				{
					Id:        uuid.UUID(gameImageID2),
					Mime:      openapi.Imagepng,
					CreatedAt: now,
				},
			},
		},
		{
			description: "gifでもエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			images: []*domain.GameImage{
				domain.NewGameImage(
					gameImageID3,
					values.GameImageTypeGif,
					now,
				),
			},
			resImages: []openapi.GameImage{
				{
					Id:        uuid.UUID(gameImageID3),
					Mime:      openapi.Imagegif,
					CreatedAt: now,
				},
			},
		},
		{
			description: "jpeg,png,gifのいずれでもないので500",
			gameID:      uuid.UUID(values.NewGameID()),
			images: []*domain.GameImage{
				domain.NewGameImage(
					values.NewGameImageID(),
					values.GameImageType(100),
					now,
				),
			},
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description:      "GetGameImagesがErrInvalidGameIDなので404",
			gameID:           uuid.UUID(values.NewGameID()),
			getGameImagesErr: service.ErrInvalidGameID,
			isErr:            true,
			statusCode:       http.StatusNotFound,
		},
		{
			description:      "GetGameImagesがエラーなので500",
			gameID:           uuid.UUID(values.NewGameID()),
			getGameImagesErr: errors.New("error"),
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
		{
			description: "画像がなくても問題なし",
			gameID:      uuid.UUID(values.NewGameID()),
			images:      []*domain.GameImage{},
			resImages:   []openapi.GameImage{},
		},
		{
			description: "画像が複数あっても問題なし",
			gameID:      uuid.UUID(values.NewGameID()),
			images: []*domain.GameImage{
				domain.NewGameImage(
					gameImageID4,
					values.GameImageTypeJpeg,
					now,
				),
				domain.NewGameImage(
					gameImageID5,
					values.GameImageTypePng,
					now.Add(-10*time.Hour),
				),
			},
			resImages: []openapi.GameImage{
				{
					Id:        uuid.UUID(gameImageID4),
					Mime:      openapi.Imagejpeg,
					CreatedAt: now,
				},
				{
					Id:        uuid.UUID(gameImageID5),
					Mime:      openapi.Imagepng,
					CreatedAt: now.Add(-10 * time.Hour),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/api/v2/games/%s/images", testCase.gameID), nil)

			mockGameImageService.
				EXPECT().
				GetGameImages(gomock.Any(), gomock.Any()).
				Return(testCase.images, testCase.getGameImagesErr)

			err := gameImage.GetGameImages(c, testCase.gameID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Equal(t, http.StatusOK, rec.Code)

			var resImages []openapi.GameImage
			err = json.NewDecoder(rec.Body).Decode(&resImages)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			for i, resImage := range resImages {
				assert.Equal(t, testCase.resImages[i].Id, resImage.Id)
				assert.Equal(t, testCase.resImages[i].Mime, resImage.Mime)
				assert.WithinDuration(t, testCase.resImages[i].CreatedAt, resImage.CreatedAt, time.Second)
			}
		})
	}
}

func TestPostGameImage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameImageService := mock.NewMockGameImageV2(ctrl)

	gameImage := NewGameImage(mockGameImageService)

	type test struct {
		description          string
		gameID               openapi.GameIDInPath
		reader               *bytes.Reader
		executeSaveGameImage bool
		image                *domain.GameImage
		saveGameImageErr     error
		resImage             openapi.GameImage
		isErr                bool
		err                  error
		statusCode           int
	}

	gameImageID1 := values.NewGameImageID()
	gameImageID2 := values.NewGameImageID()
	gameImageID3 := values.NewGameImageID()

	now := time.Now()
	testCases := []test{
		{
			description:          "特に問題ないのでエラーなし",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameImage: true,
			image: domain.NewGameImage(
				gameImageID1,
				values.GameImageTypeJpeg,
				now,
			),
			resImage: openapi.GameImage{
				Id:        uuid.UUID(gameImageID1),
				Mime:      openapi.Imagejpeg,
				CreatedAt: now,
			},
		},
		{
			description:          "pngでもエラーなし",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameImage: true,
			image: domain.NewGameImage(
				gameImageID2,
				values.GameImageTypePng,
				now,
			),
			resImage: openapi.GameImage{
				Id:        uuid.UUID(gameImageID2),
				Mime:      openapi.Imagepng,
				CreatedAt: now,
			},
		},
		{
			description:          "gifでもエラーなし",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameImage: true,
			image: domain.NewGameImage(
				gameImageID3,
				values.GameImageTypeGif,
				now,
			),
			resImage: openapi.GameImage{
				Id:        uuid.UUID(gameImageID3),
				Mime:      openapi.Imagegif,
				CreatedAt: now,
			},
		},
		{
			// serviceが正しく動作していればあり得ないが、念のため確認
			description:          "jpeg,png,gifのいずれでもないので500",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameImage: true,
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageType(100),
				now,
			),
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description:          "SaveGameImageがErrInvalidGameIDなので404",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameImage: true,
			saveGameImageErr:     service.ErrInvalidGameID,
			isErr:                true,
			statusCode:           http.StatusNotFound,
		},
		{
			description:          "SaveGameImageがエラーなので500",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameImage: true,
			saveGameImageErr:     errors.New("error"),
			isErr:                true,
			statusCode:           http.StatusInternalServerError,
		},
		{
			description: "contentがrequest bodyにないので400",
			gameID:      uuid.UUID(values.NewGameID()),
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			formDatas := []testFormData{}
			if testCase.reader != nil {
				formDatas = append(formDatas, testFormData{
					fieldName: "content",
					fileName:  "content",
					body:      testCase.reader,
					isFile:    true,
				})
			}
			c, _, rec := setupTestRequest(t, http.MethodPost, fmt.Sprintf("/api/v2/games/%s/images", testCase.gameID),
				withMultipartFormDataBody(t, formDatas))

			if testCase.executeSaveGameImage {
				mockGameImageService.
					EXPECT().
					SaveGameImage(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(testCase.image, testCase.saveGameImageErr)
			}

			err := gameImage.PostGameImage(c, testCase.gameID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Equal(t, http.StatusCreated, rec.Code)

			var resImage openapi.GameImage
			err = json.NewDecoder(rec.Body).Decode(&resImage)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			assert.Equal(t, testCase.resImage.Id, resImage.Id)
			assert.Equal(t, testCase.resImage.Mime, resImage.Mime)
			assert.WithinDuration(t, testCase.resImage.CreatedAt, resImage.CreatedAt, time.Second)
		})
	}
}

func TestGetGameImage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameImageService := mock.NewMockGameImageV2(ctrl)

	gameImage := NewGameImage(mockGameImageService)

	type test struct {
		description     string
		gameID          openapi.GameIDInPath
		gameImageID     openapi.GameImageIDInPath
		tmpURL          values.GameImageTmpURL
		getGameImageErr error
		resLocation     string
		isErr           bool
		err             error
		statusCode      int
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			gameImageID: uuid.UUID(values.NewGameImageID()),
			tmpURL:      values.NewGameImageTmpURL(urlLink),
			resLocation: "https://example.com",
		},
		{
			description:     "GetGameImageがErrInvalidGameIDなので404",
			gameID:          uuid.UUID(values.NewGameID()),
			gameImageID:     uuid.UUID(values.NewGameImageID()),
			getGameImageErr: service.ErrInvalidGameID,
			isErr:           true,
			statusCode:      http.StatusNotFound,
		},
		{
			description:     "GetGameImageがErrInvalidGameImageIDなので404",
			gameID:          uuid.UUID(values.NewGameID()),
			gameImageID:     uuid.UUID(values.NewGameImageID()),
			getGameImageErr: service.ErrInvalidGameImageID,
			isErr:           true,
			statusCode:      http.StatusNotFound,
		},
		{
			description:     "GetGameImageがエラーなので500",
			gameID:          uuid.UUID(values.NewGameID()),
			gameImageID:     uuid.UUID(values.NewGameImageID()),
			getGameImageErr: errors.New("error"),
			isErr:           true,
			statusCode:      http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/api/v2/games/%s/images", testCase.gameID), nil)

			mockGameImageService.
				EXPECT().
				GetGameImage(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(testCase.tmpURL, testCase.getGameImageErr)

			err := gameImage.GetGameImage(c, testCase.gameID, testCase.gameImageID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Equal(t, http.StatusSeeOther, rec.Code)

			assert.Equal(t, testCase.resLocation, rec.Header().Get("Location"))
		})
	}
}

func TestGetGameImageMeta(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameImageService := mock.NewMockGameImageV2(ctrl)

	gameImage := NewGameImage(mockGameImageService)

	type test struct {
		description         string
		gameID              openapi.GameIDInPath
		gameImageID         openapi.GameImageIDInPath
		image               *domain.GameImage
		getGameImageMetaErr error
		resImage            openapi.GameImage
		isErr               bool
		err                 error
		statusCode          int
	}

	gameImageID1 := values.NewGameImageID()
	gameImageID2 := values.NewGameImageID()
	gameImageID3 := values.NewGameImageID()

	now := time.Now()
	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			image: domain.NewGameImage(
				gameImageID1,
				values.GameImageTypeJpeg,
				now,
			),
			resImage: openapi.GameImage{
				Id:        uuid.UUID(gameImageID1),
				Mime:      openapi.Imagejpeg,
				CreatedAt: now,
			},
		},
		{
			description: "pngでもエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			image: domain.NewGameImage(
				gameImageID2,
				values.GameImageTypePng,
				now,
			),
			resImage: openapi.GameImage{
				Id:        uuid.UUID(gameImageID2),
				Mime:      openapi.Imagepng,
				CreatedAt: now,
			},
		},
		{
			description: "gifでもエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			image: domain.NewGameImage(
				gameImageID3,
				values.GameImageTypeGif,
				now,
			),
			resImage: openapi.GameImage{
				Id:        uuid.UUID(gameImageID3),
				Mime:      openapi.Imagegif,
				CreatedAt: now,
			},
		},
		{
			description: "jpeg,png,gifのいずれでもないので500",
			gameID:      uuid.UUID(values.NewGameID()),
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageType(100),
				now,
			),
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description:         "GetGameImageMetaがErrInvalidGameIDなので404",
			gameID:              uuid.UUID(values.NewGameID()),
			getGameImageMetaErr: service.ErrInvalidGameID,
			isErr:               true,
			statusCode:          http.StatusNotFound,
		},
		{
			description:         "GetGameImageMetaがErrInvalidGameImageIDなので404",
			gameID:              uuid.UUID(values.NewGameID()),
			getGameImageMetaErr: service.ErrInvalidGameImageID,
			isErr:               true,
			statusCode:          http.StatusNotFound,
		},
		{
			description:         "GetGameImagesがエラーなので500",
			gameID:              uuid.UUID(values.NewGameID()),
			getGameImageMetaErr: errors.New("error"),
			isErr:               true,
			statusCode:          http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/api/v2/games/%s/images", testCase.gameID), nil)

			mockGameImageService.
				EXPECT().
				GetGameImageMeta(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(testCase.image, testCase.getGameImageMetaErr)

			err := gameImage.GetGameImageMeta(c, testCase.gameID, testCase.gameImageID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Equal(t, http.StatusOK, rec.Code)

			var resImage openapi.GameImage
			err = json.NewDecoder(rec.Body).Decode(&resImage)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			assert.Equal(t, testCase.resImage.Id, resImage.Id)
			assert.Equal(t, testCase.resImage.Mime, resImage.Mime)
			assert.WithinDuration(t, testCase.resImage.CreatedAt, resImage.CreatedAt, time.Second)
		})
	}
}
