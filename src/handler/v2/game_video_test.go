package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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

func TestGetGameVideos(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameVideoService := mock.NewMockGameVideoV2(ctrl)

	gameVideo := NewGameVideo(mockGameVideoService)

	type test struct {
		description      string
		gameID           openapi.GameIDInPath
		videos           []*domain.GameVideo
		getGameVideosErr error
		resVideos        []openapi.GameVideo
		isErr            bool
		err              error
		statusCode       int
	}

	gameVideoID1 := values.NewGameVideoID()
	gameVideoID2 := values.NewGameVideoID()
	gameVideoID3 := values.NewGameVideoID()
	gameVideoID4 := values.NewGameVideoID()

	now := time.Now()
	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			videos: []*domain.GameVideo{
				domain.NewGameVideo(
					gameVideoID1,
					values.GameVideoTypeMp4,
					now,
				),
			},
			resVideos: []openapi.GameVideo{
				{
					Id:        uuid.UUID(gameVideoID1),
					Mime:      openapi.Videomp4,
					CreatedAt: now,
				},
			},
		},
		{
			description: "動画タイプがmp4,mkv,m4vでないので500",
			gameID:      uuid.UUID(values.NewGameID()),
			videos: []*domain.GameVideo{
				domain.NewGameVideo(
					values.NewGameVideoID(),
					values.GameVideoType(100),
					now,
				),
			},
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description:      "GetGameVideosがErrInvalidGameIDなので404",
			gameID:           uuid.UUID(values.NewGameID()),
			getGameVideosErr: service.ErrInvalidGameID,
			isErr:            true,
			statusCode:       http.StatusNotFound,
		},
		{
			description:      "GetGameVideosがエラーなので500",
			gameID:           uuid.UUID(values.NewGameID()),
			getGameVideosErr: errors.New("error"),
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
		{
			description: "動画がなくても問題なし",
			gameID:      uuid.UUID(values.NewGameID()),
			videos:      []*domain.GameVideo{},
			resVideos:   []openapi.GameVideo{},
		},
		{
			description: "動画が複数あっても問題なし",
			gameID:      uuid.UUID(values.NewGameID()),
			videos: []*domain.GameVideo{
				domain.NewGameVideo(
					gameVideoID2,
					values.GameVideoTypeMp4,
					now,
				),
				domain.NewGameVideo(
					gameVideoID3,
					values.GameVideoTypeMp4,
					now.Add(-10*time.Hour),
				),
			},
			resVideos: []openapi.GameVideo{
				{
					Id:        uuid.UUID(gameVideoID2),
					Mime:      openapi.Videomp4,
					CreatedAt: now,
				},
				{
					Id:        uuid.UUID(gameVideoID3),
					Mime:      openapi.Videomp4,
					CreatedAt: now.Add(-10 * time.Hour),
				},
			},
		},
		{
			description: "動画のタイプが複数あっても問題なし",
			gameID:      uuid.UUID(values.NewGameID()),
			videos: []*domain.GameVideo{
				domain.NewGameVideo(
					gameVideoID2,
					values.GameVideoTypeMp4,
					now,
				),
				domain.NewGameVideo(
					gameVideoID3,
					values.GameVideoTypeM4v,
					now.Add(-10*time.Hour),
				),
				domain.NewGameVideo(
					gameVideoID4,
					values.GameVideoTypeMkv,
					now.Add(-20*time.Hour),
				),
			},
			resVideos: []openapi.GameVideo{
				{
					Id:        uuid.UUID(gameVideoID2),
					Mime:      openapi.Videomp4,
					CreatedAt: now,
				},
				{
					Id:        uuid.UUID(gameVideoID3),
					Mime:      openapi.Videom4v,
					CreatedAt: now.Add(-10 * time.Hour),
				},
				{
					Id:        uuid.UUID(gameVideoID4),
					Mime:      openapi.Videomkv,
					CreatedAt: now.Add(-20 * time.Hour),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v2/games/%s/videos", testCase.gameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockGameVideoService.
				EXPECT().
				GetGameVideos(gomock.Any(), gomock.Any()).
				Return(testCase.videos, testCase.getGameVideosErr)

			err := gameVideo.GetGameVideos(c, testCase.gameID)

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

			var resVideos []openapi.GameVideo
			err = json.NewDecoder(rec.Body).Decode(&resVideos)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			for i, resVideo := range resVideos {
				assert.Equal(t, testCase.resVideos[i].Id, resVideo.Id)
				assert.Equal(t, testCase.resVideos[i].Mime, resVideo.Mime)
				assert.WithinDuration(t, testCase.resVideos[i].CreatedAt, resVideo.CreatedAt, time.Second)
			}
		})
	}
}

func TestPostGameVideo(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameVideoService := mock.NewMockGameVideoV2(ctrl)

	gameVideo := NewGameVideo(mockGameVideoService)

	type test struct {
		description          string
		gameID               openapi.GameIDInPath
		reader               *bytes.Reader
		executeSaveGameVideo bool
		video                *domain.GameVideo
		saveGameVideoErr     error
		resVideo             openapi.GameVideo
		isErr                bool
		err                  error
		statusCode           int
	}

	gameVideoID1 := values.NewGameVideoID()

	now := time.Now()
	testCases := []test{
		{
			description:          "特に問題ないのでエラーなし",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameVideo: true,
			video: domain.NewGameVideo(
				gameVideoID1,
				values.GameVideoTypeMp4,
				now,
			),
			resVideo: openapi.GameVideo{
				Id:        uuid.UUID(gameVideoID1),
				Mime:      openapi.Videomp4,
				CreatedAt: now,
			},
		},
		{
			description:          "m4vでもエラーなし",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameVideo: true,
			video: domain.NewGameVideo(
				gameVideoID1,
				values.GameVideoTypeM4v,
				now,
			),
			resVideo: openapi.GameVideo{
				Id:        uuid.UUID(gameVideoID1),
				Mime:      openapi.Videom4v,
				CreatedAt: now,
			},
		},
		{
			description:          "mkvでもエラーなし",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameVideo: true,
			video: domain.NewGameVideo(
				gameVideoID1,
				values.GameVideoTypeMkv,
				now,
			),
			resVideo: openapi.GameVideo{
				Id:        uuid.UUID(gameVideoID1),
				Mime:      openapi.Videomkv,
				CreatedAt: now,
			},
		},
		{
			// serviceが正しく動作していればあり得ないが、念のため確認
			description:          "mp4,m4v,mkvでないので500",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameVideo: true,
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoType(100),
				now,
			),
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description:          "SaveGameVideoがErrInvalidGameIDなので404",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameVideo: true,
			saveGameVideoErr:     service.ErrInvalidGameID,
			isErr:                true,
			statusCode:           http.StatusNotFound,
		},
		{
			description:          "SaveGameVideoがErrInvalidFormatなので400",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameVideo: true,
			saveGameVideoErr:     service.ErrInvalidFormat,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description:          "SaveGameVideoがエラーなので500",
			gameID:               uuid.UUID(values.NewGameID()),
			reader:               bytes.NewReader([]byte("test")),
			executeSaveGameVideo: true,
			saveGameVideoErr:     errors.New("error"),
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
			e := echo.New()

			reqBody := bytes.NewBuffer(nil)
			var boundary string
			func() {
				mw := multipart.NewWriter(reqBody)
				defer mw.Close()

				if testCase.reader != nil {
					w, err := mw.CreateFormFile("content", "content")
					if err != nil {
						t.Fatalf("failed to create form field: %v", err)
						return
					}

					_, err = io.Copy(w, testCase.reader)
					if err != nil {
						t.Fatalf("failed to copy: %v", err)
						return
					}
				}

				boundary = mw.Boundary()
			}()

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v2/games/%s/videos", testCase.gameID), reqBody)
			req.Header.Set(echo.HeaderContentType, fmt.Sprintf("multipart/form-data; boundary=%s", boundary))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeSaveGameVideo {
				mockGameVideoService.
					EXPECT().
					SaveGameVideo(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(testCase.video, testCase.saveGameVideoErr)
			}

			err := gameVideo.PostGameVideo(c, testCase.gameID)

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

			var resVideo openapi.GameVideo
			err = json.NewDecoder(rec.Body).Decode(&resVideo)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			assert.Equal(t, testCase.resVideo.Id, resVideo.Id)
			assert.Equal(t, testCase.resVideo.Mime, resVideo.Mime)
			assert.WithinDuration(t, testCase.resVideo.CreatedAt, resVideo.CreatedAt, time.Second)
		})
	}
}

func TestGetGameVideo(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameVideoService := mock.NewMockGameVideoV2(ctrl)

	gameVideo := NewGameVideo(mockGameVideoService)

	type test struct {
		description     string
		gameID          openapi.GameIDInPath
		gameVideoID     openapi.GameVideoIDInPath
		tmpURL          values.GameVideoTmpURL
		getGameVideoErr error
		resLocation     string
		isErr           bool
		err             error
		statusCode      int
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode video: %v", err)
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			gameVideoID: uuid.UUID(values.NewGameVideoID()),
			tmpURL:      values.NewGameVideoTmpURL(urlLink),
			resLocation: "https://example.com",
		},
		{
			description:     "GetGameVideoがErrInvalidGameIDなので404",
			gameID:          uuid.UUID(values.NewGameID()),
			gameVideoID:     uuid.UUID(values.NewGameVideoID()),
			getGameVideoErr: service.ErrInvalidGameID,
			isErr:           true,
			statusCode:      http.StatusNotFound,
		},
		{
			description:     "GetGameVideoがErrInvalidGameVideoIDなので404",
			gameID:          uuid.UUID(values.NewGameID()),
			gameVideoID:     uuid.UUID(values.NewGameVideoID()),
			getGameVideoErr: service.ErrInvalidGameVideoID,
			isErr:           true,
			statusCode:      http.StatusNotFound,
		},
		{
			description:     "GetGameVideoがエラーなので500",
			gameID:          uuid.UUID(values.NewGameID()),
			gameVideoID:     uuid.UUID(values.NewGameVideoID()),
			getGameVideoErr: errors.New("error"),
			isErr:           true,
			statusCode:      http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v2/games/%s/videos", testCase.gameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockGameVideoService.
				EXPECT().
				GetGameVideo(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(testCase.tmpURL, testCase.getGameVideoErr)

			err := gameVideo.GetGameVideo(c, testCase.gameID, testCase.gameVideoID)

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

func TestGetGameVideoMeta(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameVideoService := mock.NewMockGameVideoV2(ctrl)

	gameVideo := NewGameVideo(mockGameVideoService)

	type test struct {
		description         string
		gameID              openapi.GameIDInPath
		gameVideoID         openapi.GameVideoIDInPath
		video               *domain.GameVideo
		getGameVideoMetaErr error
		resVideo            openapi.GameVideo
		isErr               bool
		err                 error
		statusCode          int
	}

	gameVideoID1 := values.NewGameVideoID()

	now := time.Now()
	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			video: domain.NewGameVideo(
				gameVideoID1,
				values.GameVideoTypeMp4,
				now,
			),
			resVideo: openapi.GameVideo{
				Id:        uuid.UUID(gameVideoID1),
				Mime:      openapi.Videomp4,
				CreatedAt: now,
			},
		},
		{
			description: "m4vで問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			video: domain.NewGameVideo(
				gameVideoID1,
				values.GameVideoTypeM4v,
				now,
			),
			resVideo: openapi.GameVideo{
				Id:        uuid.UUID(gameVideoID1),
				Mime:      openapi.Videom4v,
				CreatedAt: now,
			},
		},
		{
			description: "mkvで問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			video: domain.NewGameVideo(
				gameVideoID1,
				values.GameVideoTypeMkv,
				now,
			),
			resVideo: openapi.GameVideo{
				Id:        uuid.UUID(gameVideoID1),
				Mime:      openapi.Videomkv,
				CreatedAt: now,
			},
		},
		{
			description: "mp4,mkv,m4vでないので500",
			gameID:      uuid.UUID(values.NewGameID()),
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoType(100),
				now,
			),
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description:         "GetGameVideoMetaがErrInvalidGameIDなので404",
			gameID:              uuid.UUID(values.NewGameID()),
			getGameVideoMetaErr: service.ErrInvalidGameID,
			isErr:               true,
			statusCode:          http.StatusNotFound,
		},
		{
			description:         "GetGameVideoMetaがErrInvalidGameVideoIDなので404",
			gameID:              uuid.UUID(values.NewGameID()),
			getGameVideoMetaErr: service.ErrInvalidGameVideoID,
			isErr:               true,
			statusCode:          http.StatusNotFound,
		},
		{
			description:         "GetGameVideosがエラーなので500",
			gameID:              uuid.UUID(values.NewGameID()),
			getGameVideoMetaErr: errors.New("error"),
			isErr:               true,
			statusCode:          http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v2/games/%s/videos", testCase.gameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockGameVideoService.
				EXPECT().
				GetGameVideoMeta(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(testCase.video, testCase.getGameVideoMetaErr)

			err := gameVideo.GetGameVideoMeta(c, testCase.gameID, testCase.gameVideoID)

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

			var resVideo openapi.GameVideo
			err = json.NewDecoder(rec.Body).Decode(&resVideo)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			assert.Equal(t, testCase.resVideo.Id, resVideo.Id)
			assert.Equal(t, testCase.resVideo.Mime, resVideo.Mime)
			assert.WithinDuration(t, testCase.resVideo.CreatedAt, resVideo.CreatedAt, time.Second)
		})
	}
}
