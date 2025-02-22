package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/traPtitech/trap-collection-server/pkg/types"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestGetEditions(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEditionService := mock.NewMockEdition(ctrl)
	edition := NewEdition(mockEditionService)

	type test struct {
		description      string
		editions         []*domain.LauncherVersion
		getEditionsErr   error
		expectedEditions []openapi.Edition
		isErr            bool
		statusCode       int
	}

	now := time.Now()
	editionID1 := values.NewLauncherVersionIDFromUUID(uuid.New())
	editionID2 := values.NewLauncherVersionIDFromUUID(uuid.New())
	editionName1 := values.NewLauncherVersionName("テストエディション")
	editionName2 := values.NewLauncherVersionName("テストエディション2")
	strURL := "https://example.com/questionnaire"
	questionnaireURL, err := url.Parse(strURL)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			editions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithQuestionnaire(
					editionID1,
					editionName1,
					values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
					now,
				),
			},
			expectedEditions: []openapi.Edition{
				{
					Id:            uuid.UUID(editionID1),
					Name:          string(editionName1),
					Questionnaire: &strURL,
					CreatedAt:     now,
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "アンケートURLが無くてもエラーなし",
			editions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithoutQuestionnaire(
					editionID1,
					editionName1,
					now,
				),
			},
			expectedEditions: []openapi.Edition{
				{
					Id:            uuid.UUID(editionID1),
					Name:          string(editionName1),
					Questionnaire: nil,
					CreatedAt:     now,
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description:    "GetEditionsがエラーなので500",
			getEditionsErr: errors.New("error"),
			isErr:          true,
			statusCode:     http.StatusInternalServerError,
		},
		{
			description: "複数エディションでもエラーなし",
			editions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithQuestionnaire(
					editionID1,
					editionName1,
					values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
					now,
				),
				domain.NewLauncherVersionWithoutQuestionnaire(
					editionID2,
					editionName2,
					now,
				),
			},
			expectedEditions: []openapi.Edition{
				{
					Id:            uuid.UUID(editionID1),
					Name:          string(editionName1),
					Questionnaire: &strURL,
					CreatedAt:     now,
				},
				{
					Id:            uuid.UUID(editionID2),
					Name:          string(editionName2),
					Questionnaire: nil,
					CreatedAt:     now,
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "エディションが存在しなくてもでもエラーなし",
			editions:    []*domain.LauncherVersion{},
			statusCode:  http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v2/editions", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockEditionService.
				EXPECT().
				GetEditions(gomock.Any()).
				Return(testCase.editions, testCase.getEditionsErr)

			err := edition.GetEditions(c)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if errors.As(err, &httpErr) {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					} else {
						t.Errorf("error should be *echo.HTTPError, but got %T", err)
					}
				} else {
					assert.Error(t, err)
				}
				return
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, testCase.statusCode, rec.Code)

			var res []openapi.Edition
			err = json.NewDecoder(rec.Body).Decode(&res)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Len(t, res, len(testCase.expectedEditions))
			for i, ed := range res {
				assert.Equal(t, testCase.expectedEditions[i].Id, ed.Id)
				assert.Equal(t, testCase.expectedEditions[i].Name, ed.Name)
				assert.Equal(t, testCase.expectedEditions[i].Questionnaire, ed.Questionnaire)
				assert.WithinDuration(t, testCase.expectedEditions[i].CreatedAt, ed.CreatedAt, 2*time.Second)
			}
		})
	}

}

func TestPostEdition(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEditionService := mock.NewMockEdition(ctrl)
	edition := NewEdition(mockEditionService)

	type test struct {
		description          string
		reqBody              *openapi.NewEdition
		invalidBody          bool
		executeCreateEdition bool
		name                 values.LauncherVersionName
		questionnaireURL     types.Option[values.LauncherVersionQuestionnaireURL]
		gameVersionIDs       []values.GameVersionID
		createEditionErr     error
		resultEdition        *domain.LauncherVersion
		isErr                bool
		statusCode           int
		expectEdition        *openapi.Edition
	}

	now := time.Now()
	editionUUID := uuid.New()
	editionID := values.NewLauncherVersionIDFromUUID(editionUUID)
	editionName := "テストエディション"
	strURL := "https://example.com/questionnaire"
	invalidURL := " https://example.com/questionnaire"
	longName := strings.Repeat("あ", 33)
	questionnaireURL, err := url.Parse(strURL)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}
	gameVersionUUID1 := uuid.New()
	gameVersionUUID2 := uuid.New()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			reqBody: &openapi.NewEdition{
				Name:          editionName,
				Questionnaire: &strURL,
				GameVersions:  []uuid.UUID{gameVersionUUID1, gameVersionUUID2},
			},
			executeCreateEdition: true,
			name:                 values.NewLauncherVersionName(editionName),
			questionnaireURL:     types.NewOption(values.NewLauncherVersionQuestionnaireURL(questionnaireURL)),
			gameVersionIDs: []values.GameVersionID{
				values.NewGameVersionIDFromUUID(gameVersionUUID1),
				values.NewGameVersionIDFromUUID(gameVersionUUID2),
			},
			resultEdition: domain.NewLauncherVersionWithQuestionnaire(
				editionID,
				values.NewLauncherVersionName(editionName),
				values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
				now,
			),
			expectEdition: &openapi.Edition{
				Id:            editionUUID,
				Name:          editionName,
				Questionnaire: &strURL,
				CreatedAt:     now,
			},
			statusCode: http.StatusCreated,
		},
		{
			description: "アンケートURLがなくてもエラーなし",
			reqBody: &openapi.NewEdition{
				Name:         editionName,
				GameVersions: []uuid.UUID{gameVersionUUID1},
			},
			executeCreateEdition: true,
			name:                 values.NewLauncherVersionName(editionName),
			questionnaireURL:     types.Option[values.LauncherVersionQuestionnaireURL]{},
			gameVersionIDs:       []values.GameVersionID{values.NewGameVersionIDFromUUID(gameVersionUUID1)},
			resultEdition: domain.NewLauncherVersionWithoutQuestionnaire(
				editionID,
				values.NewLauncherVersionName(editionName),
				now,
			),
			expectEdition: &openapi.Edition{
				Id:            editionUUID,
				Name:          editionName,
				Questionnaire: nil,
				CreatedAt:     now,
			},
			statusCode: http.StatusCreated,
		},
		{
			description: "Edition名が空文字なので400",
			reqBody: &openapi.NewEdition{
				Name:         "",
				GameVersions: []uuid.UUID{gameVersionUUID1},
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "Edition名が長すぎるので400",
			reqBody: &openapi.NewEdition{
				Name:         longName,
				GameVersions: []uuid.UUID{gameVersionUUID1},
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "URLが正しくないので400",
			reqBody: &openapi.NewEdition{
				Name:          editionName,
				Questionnaire: &invalidURL,
				GameVersions:  []uuid.UUID{gameVersionUUID1},
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "ゲームバージョンが重複しているので400",
			reqBody: &openapi.NewEdition{
				Name:         editionName,
				GameVersions: []uuid.UUID{gameVersionUUID1, gameVersionUUID1},
			},
			executeCreateEdition: true,
			name:                 values.NewLauncherVersionName(editionName),
			questionnaireURL:     types.Option[values.LauncherVersionQuestionnaireURL]{},
			gameVersionIDs: []values.GameVersionID{
				values.NewGameVersionIDFromUUID(gameVersionUUID1),
				values.NewGameVersionIDFromUUID(gameVersionUUID1),
			},
			createEditionErr: service.ErrDuplicateGameVersion,
			isErr:            true,
			statusCode:       http.StatusBadRequest,
		},
		{
			description: "無効なゲームバージョンIDが含まれているので400",
			reqBody: &openapi.NewEdition{
				Name:         editionName,
				GameVersions: []uuid.UUID{gameVersionUUID1},
			},
			executeCreateEdition: true,
			name:                 values.NewLauncherVersionName(editionName),
			questionnaireURL:     types.Option[values.LauncherVersionQuestionnaireURL]{},
			gameVersionIDs:       []values.GameVersionID{values.NewGameVersionIDFromUUID(gameVersionUUID1)},
			createEditionErr:     service.ErrInvalidGameVersionID,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description: "サービス層でエラーが発生したので500",
			reqBody: &openapi.NewEdition{
				Name:         editionName,
				GameVersions: []uuid.UUID{gameVersionUUID1},
			},
			executeCreateEdition: true,
			name:                 values.NewLauncherVersionName(editionName),
			questionnaireURL:     types.Option[values.LauncherVersionQuestionnaireURL]{},
			gameVersionIDs:       []values.GameVersionID{values.NewGameVersionIDFromUUID(gameVersionUUID1)},
			createEditionErr:     errors.New("internal error"),
			isErr:                true,
			statusCode:           http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			var req *http.Request
			if testCase.invalidBody {
				reqBody := bytes.NewBuffer([]byte("invalid"))
				req = httptest.NewRequest(http.MethodPost, "/api/v2/editions", reqBody)
				req.Header.Set("Content-Type", echo.MIMETextPlain)
			} else {
				reqBody := bytes.NewBuffer(nil)
				if err := json.NewEncoder(reqBody).Encode(testCase.reqBody); err != nil {
					t.Fatalf("failed to encode request body: %v", err)
				}
				req = httptest.NewRequest(http.MethodPost, "/api/v2/editions", reqBody)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeCreateEdition {
				mockEditionService.
					EXPECT().
					CreateEdition(
						gomock.Any(),
						testCase.name,
						testCase.questionnaireURL,
						testCase.gameVersionIDs,
					).
					Return(testCase.resultEdition, testCase.createEditionErr)
			}

			err := edition.PostEdition(c)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if errors.As(err, &httpErr) {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else {
					assert.Error(t, err)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)

			if testCase.expectEdition != nil {
				var res openapi.Edition
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				assert.Equal(t, testCase.expectEdition.Id, res.Id)
				assert.Equal(t, testCase.expectEdition.Name, res.Name)
				assert.Equal(t, testCase.expectEdition.Questionnaire, res.Questionnaire)
				assert.WithinDuration(t, testCase.expectEdition.CreatedAt, res.CreatedAt, time.Second)
			}
		})
	}
}

func TestDeleteEdition(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEditionService := mock.NewMockEdition(ctrl)
	edition := NewEdition(mockEditionService)

	type test struct {
		editionID        openapi.EditionIDInPath
		deleteEditionErr error
		isErr            bool
		statusCode       int
	}

	testCases := map[string]test{
		"正常系:特に問題ないのでエラー無し": {
			editionID: uuid.New(),
		},
		"異常系:存在しないエディションIDなので400": {
			editionID:        uuid.New(),
			deleteEditionErr: service.ErrInvalidEditionID,
			isErr:            true,
			statusCode:       http.StatusBadRequest,
		},
		"異常系:DeleteEditionがエラーなのでエラー": {
			editionID:        uuid.New(),
			deleteEditionErr: errors.New("internal error"),
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			mockEditionService.
				EXPECT().
				DeleteEdition(gomock.Any(), values.NewLauncherVersionIDFromUUID(testCase.editionID)).
				Return(testCase.deleteEditionErr)

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v2/editions/%s", testCase.editionID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := edition.DeleteEdition(c, testCase.editionID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if errors.As(err, &httpErr) {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					} else {
						t.Errorf("error should be *echo.HTTPError, but got %T", err)
					}
				} else {
					assert.Error(t, err)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestGetEdition(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEditionService := mock.NewMockEdition(ctrl)
	edition := NewEdition(mockEditionService)

	type test struct {
		editionID     openapi.EditionIDInPath
		getEditionErr error
		resultEdition *domain.LauncherVersion
		isErr         bool
		statusCode    int
		expectedRes   *openapi.Edition
	}

	now := time.Now()
	editionUUID := uuid.New()
	editionID := values.NewLauncherVersionIDFromUUID(editionUUID)
	editionName := values.NewLauncherVersionName("テストエディション")
	questionnaireURL, _ := url.Parse("https://example.com/questionnaire")

	testCases := map[string]test{
		"正常系:アンケートURLありのエディションが取得できる": {
			editionID: editionUUID,
			resultEdition: domain.NewLauncherVersionWithQuestionnaire(
				editionID,
				editionName,
				values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
				now,
			),
			expectedRes: &openapi.Edition{
				Id:            editionUUID,
				Name:          string(editionName),
				Questionnaire: ptr(questionnaireURL.String()),
				CreatedAt:     now,
			},
			statusCode: http.StatusOK,
		},
		"正常系:アンケートURLなしのエディションが取得できる": {
			editionID: editionUUID,
			resultEdition: domain.NewLauncherVersionWithoutQuestionnaire(
				editionID,
				editionName,
				now,
			),
			expectedRes: &openapi.Edition{
				Id:            editionUUID,
				Name:          string(editionName),
				Questionnaire: nil,
				CreatedAt:     now,
			},
			statusCode: http.StatusOK,
		},
		"異常系:存在しないエディションIDなので400": {
			editionID:     editionUUID,
			getEditionErr: service.ErrInvalidEditionID,
			isErr:         true,
			statusCode:    http.StatusBadRequest,
		},
		"異常系:GetEditionがエラーなのでエラー": {
			editionID:     editionUUID,
			getEditionErr: errors.New("internal error"),
			isErr:         true,
			statusCode:    http.StatusInternalServerError,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			mockEditionService.
				EXPECT().
				GetEdition(gomock.Any(), values.NewLauncherVersionIDFromUUID(testCase.editionID)).
				Return(testCase.resultEdition, testCase.getEditionErr)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v2/editions/%s", testCase.editionID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := edition.GetEdition(c, testCase.editionID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if errors.As(err, &httpErr) {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					} else {
						t.Errorf("error should be *echo.HTTPError, but got %T", err)
					}
				} else {
					assert.Error(t, err)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)

			var res openapi.Edition
			err = json.NewDecoder(rec.Body).Decode(&res)
			assert.NoError(t, err)

			assert.Equal(t, testCase.expectedRes.Id, res.Id)
			assert.Equal(t, testCase.expectedRes.Name, res.Name)
			assert.Equal(t, testCase.expectedRes.Questionnaire, res.Questionnaire)
			assert.WithinDuration(t, testCase.expectedRes.CreatedAt, res.CreatedAt, time.Second)
		})
	}
}

func ptr(s string) *string {
	return &s
}

func TestPatchEdition(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEditionService := mock.NewMockEdition(ctrl)
	edition := NewEdition(mockEditionService)

	type test struct {
		editionID        openapi.EditionIDInPath
		reqBody          *openapi.PatchEdition
		invalidBody      bool
		updateEditionErr error
		resultEdition    *domain.LauncherVersion
		isErr            bool
		statusCode       int
		expectedRes      *openapi.Edition
	}

	now := time.Now()
	editionUUID := uuid.New()
	editionID := values.NewLauncherVersionIDFromUUID(editionUUID)
	editionName := "テストエディション"
	questionnaireURL := "https://example.com/questionnaire"
	parsedURL, _ := url.Parse(questionnaireURL)

	testCases := map[string]test{
		"正常系:エディションが更新できる": {
			editionID: editionUUID,
			reqBody: &openapi.PatchEdition{
				Name:          editionName,
				Questionnaire: &questionnaireURL,
			},
			resultEdition: domain.NewLauncherVersionWithQuestionnaire(
				editionID,
				values.NewLauncherVersionName(editionName),
				values.NewLauncherVersionQuestionnaireURL(parsedURL),
				now,
			),
			expectedRes: &openapi.Edition{
				Id:            editionUUID,
				Name:          editionName,
				Questionnaire: &questionnaireURL,
				CreatedAt:     now,
			},
			statusCode: http.StatusOK,
		},
		"異常系:不正なリクエストボディ": {
			editionID:   editionUUID,
			invalidBody: true,
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		"異常系:存在しないエディションID": {
			editionID: editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: editionName,
			},
			updateEditionErr: service.ErrInvalidEditionID,
			isErr:            true,
			statusCode:       http.StatusBadRequest,
		},
		"異常系:重複したゲームバージョン": {
			editionID: editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: editionName,
			},
			updateEditionErr: service.ErrDuplicateGameVersion,
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
		"異常系:重複したゲーム": {
			editionID: editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: editionName,
			},
			updateEditionErr: service.ErrDuplicateGame,
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
		"異常系:サービス層でエラー": {
			editionID: editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: editionName,
			},
			updateEditionErr: errors.New("internal error"),
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			if !testCase.invalidBody {
				mockEditionService.
					EXPECT().
					UpdateEdition(
						gomock.Any(),
						values.NewLauncherVersionIDFromUUID(testCase.editionID),
						values.NewLauncherVersionName(testCase.reqBody.Name),
						gomock.Any(),
					).
					Return(testCase.resultEdition, testCase.updateEditionErr)
			}

			var reqBody []byte
			var err error
			if !testCase.invalidBody {
				reqBody, err = json.Marshal(testCase.reqBody)
				assert.NoError(t, err)
			} else {
				reqBody = []byte("invalid json")
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v2/editions/%s", testCase.editionID), bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err = edition.PatchEdition(c, testCase.editionID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if errors.As(err, &httpErr) {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					} else {
						t.Errorf("error should be *echo.HTTPError, but got %T", err)
					}
				} else {
					assert.Error(t, err)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)

			if testCase.expectedRes != nil {
				var res openapi.Edition
				err = json.NewDecoder(rec.Body).Decode(&res)
				assert.NoError(t, err)

				assert.Equal(t, testCase.expectedRes.Id, res.Id)
				assert.Equal(t, testCase.expectedRes.Name, res.Name)
				assert.Equal(t, testCase.expectedRes.Questionnaire, res.Questionnaire)
				assert.WithinDuration(t, testCase.expectedRes.CreatedAt, res.CreatedAt, time.Second)
			}
		})
	}
}
