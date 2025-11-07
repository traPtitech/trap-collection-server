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

	"github.com/traPtitech/trap-collection-server/pkg/option"
	"go.uber.org/mock/gomock"

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

	type test struct {
		description    string
		editions       []*domain.LauncherVersion
		getEditionsErr error
		expectEditions []openapi.Edition
		isErr          bool
		statusCode     int
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
			expectEditions: []openapi.Edition{
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
			expectEditions: []openapi.Edition{
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
			expectEditions: []openapi.Edition{
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
		testCase := testCase
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockEditionService := mock.NewMockEdition(ctrl)
			edition := NewEdition(mockEditionService)

			c, _, rec := setupTestRequest(t, http.MethodGet, "/api/v2/editions", nil)

			mockEditionService.
				EXPECT().
				GetEditions(gomock.Any()).
				Return(testCase.editions, testCase.getEditionsErr)

			err := edition.GetEditions(c)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if assert.ErrorAs(t, err, &httpErr, "error should be *echo.HTTPError") {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					}
				} else {
					assert.Error(t, err)
				}
				return
			}
			assert.NoError(t, err)
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, testCase.statusCode, rec.Code)

			var res []openapi.Edition
			err = json.NewDecoder(rec.Body).Decode(&res)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Len(t, res, len(testCase.expectEditions))
			for i, ed := range res {
				assert.Equal(t, testCase.expectEditions[i].Id, ed.Id)
				assert.Equal(t, testCase.expectEditions[i].Name, ed.Name)
				assert.Equal(t, testCase.expectEditions[i].Questionnaire, ed.Questionnaire)
				assert.WithinDuration(t, testCase.expectEditions[i].CreatedAt, ed.CreatedAt, time.Second)
			}
		})
	}
}

func TestPostEdition(t *testing.T) {
	t.Parallel()

	type test struct {
		description          string
		reqBody              *openapi.NewEdition
		invalidBody          bool
		executeCreateEdition bool
		name                 values.LauncherVersionName
		questionnaireURL     option.Option[values.LauncherVersionQuestionnaireURL]
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
	invalidURL := " https://example.com/questionnaire with spaces"
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
			questionnaireURL:     option.NewOption(values.NewLauncherVersionQuestionnaireURL(questionnaireURL)),
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
			questionnaireURL:     option.Option[values.LauncherVersionQuestionnaireURL]{},
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
			questionnaireURL:     option.Option[values.LauncherVersionQuestionnaireURL]{},
			gameVersionIDs: []values.GameVersionID{
				values.NewGameVersionIDFromUUID(gameVersionUUID1),
				values.NewGameVersionIDFromUUID(gameVersionUUID1),
			},
			createEditionErr: service.ErrDuplicateGameVersion,
			isErr:            true,
			statusCode:       http.StatusBadRequest,
		},
		{
			description: "ゲームが重複しているので400",
			reqBody: &openapi.NewEdition{
				Name:         editionName,
				GameVersions: []uuid.UUID{gameVersionUUID1, gameVersionUUID2},
			},
			executeCreateEdition: true,
			name:                 values.NewLauncherVersionName(editionName),
			questionnaireURL:     option.Option[values.LauncherVersionQuestionnaireURL]{},
			gameVersionIDs: []values.GameVersionID{
				values.NewGameVersionIDFromUUID(gameVersionUUID1),
				values.NewGameVersionIDFromUUID(gameVersionUUID2),
			},
			createEditionErr: service.ErrDuplicateGame,
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
			questionnaireURL:     option.Option[values.LauncherVersionQuestionnaireURL]{},
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
			questionnaireURL:     option.Option[values.LauncherVersionQuestionnaireURL]{},
			gameVersionIDs:       []values.GameVersionID{values.NewGameVersionIDFromUUID(gameVersionUUID1)},
			createEditionErr:     errors.New("internal error"),
			isErr:                true,
			statusCode:           http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockEditionService := mock.NewMockEdition(ctrl)
			edition := NewEdition(mockEditionService)

			var c echo.Context
			var rec *httptest.ResponseRecorder
			if testCase.invalidBody {
				c, _, rec = setupTestRequest(t, http.MethodPost, "/api/v2/editions", withStringBody(t, "invalid"))
			} else {
				c, _, rec = setupTestRequest(t, http.MethodPost, "/api/v2/editions", withJSONBody(t, testCase.reqBody))
			}

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
			t.Logf("テストケース: %s", testCase.description)
			t.Logf("レスポンスコード: %d", rec.Code)

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

	editionID := uuid.New()

	type test struct {
		description       string
		editionID         openapi.EditionIDInPath
		executeDeleteMock bool
		launcherVersionID values.LauncherVersionID
		deleteEditionErr  error
		isErr             bool
		statusCode        int
	}

	testCases := []test{
		{
			description:       "特に問題ないのでエラー無し",
			editionID:         editionID,
			executeDeleteMock: true,
			launcherVersionID: values.NewLauncherVersionIDFromUUID(editionID),
			statusCode:        http.StatusOK,
		},
		{
			description:       "存在しないエディションIDなので400",
			editionID:         editionID,
			executeDeleteMock: true,
			launcherVersionID: values.NewLauncherVersionIDFromUUID(editionID),
			deleteEditionErr:  service.ErrInvalidEditionID,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:       "DeleteEditionがエラーなので500",
			editionID:         editionID,
			executeDeleteMock: true,
			launcherVersionID: values.NewLauncherVersionIDFromUUID(editionID),
			deleteEditionErr:  errors.New("internal error"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockEditionService := mock.NewMockEdition(ctrl)
			edition := NewEdition(mockEditionService)

			if testCase.executeDeleteMock {
				mockEditionService.
					EXPECT().
					DeleteEdition(gomock.Any(), testCase.launcherVersionID).
					Return(testCase.deleteEditionErr)
			}

			c, _, rec := setupTestRequest(t, http.MethodDelete, fmt.Sprintf("/api/v2/editions/%s", testCase.editionID), nil)

			err := edition.DeleteEdition(c, testCase.editionID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if assert.ErrorAs(t, err, &httpErr, "error should be *echo.HTTPError") {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
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

	type test struct {
		description   string
		editionID     openapi.EditionIDInPath
		resultEdition *domain.LauncherVersion
		GetEditionErr error
		expectEdition *openapi.Edition
		isErr         bool
		statusCode    int
	}

	now := time.Now()
	editionUUID := uuid.New()
	editionID := values.NewLauncherVersionIDFromUUID(editionUUID)
	editionName := values.NewLauncherVersionName("テストエディション")
	strURL := "https://example.com/questionnaire"
	questionnaireURL, err := url.Parse(strURL)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	testCases := []test{
		{
			description: "アンケートURLありのエディションが取得できる",
			editionID:   editionUUID,
			resultEdition: domain.NewLauncherVersionWithQuestionnaire(
				editionID,
				editionName,
				values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
				now,
			),
			expectEdition: &openapi.Edition{
				Id:            editionUUID,
				Name:          string(editionName),
				Questionnaire: &strURL,
				CreatedAt:     now,
			},
			statusCode: http.StatusOK,
		},
		{
			description: "アンケートURLなしのエディションが取得できる",
			editionID:   editionUUID,
			resultEdition: domain.NewLauncherVersionWithoutQuestionnaire(
				editionID,
				editionName,
				now,
			),
			expectEdition: &openapi.Edition{
				Id:            editionUUID,
				Name:          string(editionName),
				Questionnaire: nil,
				CreatedAt:     now,
			},
			statusCode: http.StatusOK,
		},
		{
			description:   "存在しないエディションIDなので400",
			editionID:     editionUUID,
			GetEditionErr: service.ErrInvalidEditionID,
			isErr:         true,
			statusCode:    http.StatusBadRequest,
		},
		{
			description:   "GetEditionがエラーなので500",
			editionID:     editionUUID,
			GetEditionErr: errors.New("internal error"),
			isErr:         true,
			statusCode:    http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockEditionService := mock.NewMockEdition(ctrl)
			edition := NewEdition(mockEditionService)

			mockEditionService.
				EXPECT().
				GetEdition(gomock.Any(), values.NewLauncherVersionIDFromUUID(testCase.editionID)).
				Return(testCase.resultEdition, testCase.GetEditionErr)

			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/api/v2/editions/%s", testCase.editionID), nil)

			err := edition.GetEdition(c, testCase.editionID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if assert.ErrorAs(t, err, &httpErr, "error should be *echo.HTTPError") {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
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

			assert.Equal(t, testCase.expectEdition.Id, res.Id)
			assert.Equal(t, testCase.expectEdition.Name, res.Name)
			assert.Equal(t, testCase.expectEdition.Questionnaire, res.Questionnaire)
			assert.WithinDuration(t, testCase.expectEdition.CreatedAt, res.CreatedAt, time.Second)
		})
	}
}

func TestPatchEdition(t *testing.T) {
	t.Parallel()

	type test struct {
		description       string
		editionID         openapi.EditionIDInPath
		reqBody           *openapi.PatchEdition
		invalidBody       bool
		executeUpdateMock bool
		launcherVersionID values.LauncherVersionID
		name              values.LauncherVersionName
		questionnaireURL  option.Option[values.LauncherVersionQuestionnaireURL]
		updateEditionErr  error
		resultEdition     *domain.LauncherVersion
		isErr             bool
		statusCode        int
		expectedRes       *openapi.Edition
	}

	now := time.Now()
	editionUUID := uuid.New()
	editionID := values.NewLauncherVersionIDFromUUID(editionUUID)
	editionName := "テストエディション"
	strURL := "https://example.com/questionnaire"
	invalidURL := " https://example.com/questionnaire with spaces"
	longName := strings.Repeat("あ", 33)
	questionnaireURL, err := url.Parse(strURL)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEdition{
				Name:          editionName,
				Questionnaire: &strURL,
			},
			executeUpdateMock: true,
			launcherVersionID: editionID,
			name:              values.NewLauncherVersionName(editionName),
			questionnaireURL:  option.NewOption(values.NewLauncherVersionQuestionnaireURL(questionnaireURL)),
			resultEdition: domain.NewLauncherVersionWithQuestionnaire(
				editionID,
				values.NewLauncherVersionName(editionName),
				values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
				now,
			),
			expectedRes: &openapi.Edition{
				Id:            editionUUID,
				Name:          editionName,
				Questionnaire: &strURL,
				CreatedAt:     now,
			},
			statusCode: http.StatusOK,
		},
		{
			description: "URLがなくてもエラーなし",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: editionName,
			},
			executeUpdateMock: true,
			launcherVersionID: editionID,
			name:              values.NewLauncherVersionName(editionName),
			questionnaireURL:  option.Option[values.LauncherVersionQuestionnaireURL]{},
			resultEdition: domain.NewLauncherVersionWithoutQuestionnaire(
				editionID,
				values.NewLauncherVersionName(editionName),
				now,
			),
			expectedRes: &openapi.Edition{
				Id:            editionUUID,
				Name:          editionName,
				Questionnaire: nil,
				CreatedAt:     now,
			},
			statusCode: http.StatusOK,
		},
		{
			description: "リクエストボディが不正なので400",
			editionID:   editionUUID,
			invalidBody: true,
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description: "名前が空文字なので400",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: "",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "名前が長すぎるので400",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: longName,
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "URLが正しくないので400",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEdition{
				Name:          editionName,
				Questionnaire: &invalidURL,
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "ErrInvalidEditionIDなので400",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: editionName,
			},
			executeUpdateMock: true,
			launcherVersionID: editionID,
			name:              values.NewLauncherVersionName(editionName),
			questionnaireURL:  option.Option[values.LauncherVersionQuestionnaireURL]{},
			updateEditionErr:  service.ErrInvalidEditionID,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description: "ErrDuplicateGameVersionなので500",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: editionName,
			},
			executeUpdateMock: true,
			launcherVersionID: editionID,
			name:              values.NewLauncherVersionName(editionName),
			questionnaireURL:  option.Option[values.LauncherVersionQuestionnaireURL]{},
			updateEditionErr:  service.ErrDuplicateGameVersion,
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
		{
			description: "ErrDuplicateGameなので500",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: editionName,
			},
			executeUpdateMock: true,
			launcherVersionID: editionID,
			name:              values.NewLauncherVersionName(editionName),
			questionnaireURL:  option.Option[values.LauncherVersionQuestionnaireURL]{},
			updateEditionErr:  service.ErrDuplicateGame,
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
		{
			description: "サービス層でエラーなので500",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEdition{
				Name: editionName,
			},
			executeUpdateMock: true,
			launcherVersionID: editionID,
			name:              values.NewLauncherVersionName(editionName),
			questionnaireURL:  option.Option[values.LauncherVersionQuestionnaireURL]{},
			updateEditionErr:  errors.New("internal error"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockEditionService := mock.NewMockEdition(ctrl)
			edition := NewEdition(mockEditionService)

			if !testCase.invalidBody && testCase.executeUpdateMock {
				mockEditionService.
					EXPECT().
					UpdateEdition(
						gomock.Any(),
						testCase.launcherVersionID,
						testCase.name,
						testCase.questionnaireURL,
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

			c, _, rec := setupTestRequest(t, http.MethodPatch, fmt.Sprintf("/api/v2/editions/%s", testCase.editionID), withReaderBody(t, bytes.NewReader(reqBody), echo.MIMEApplicationJSON))

			err = edition.PatchEdition(c, testCase.editionID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if assert.ErrorAs(t, err, &httpErr, "error should be *echo.HTTPError") {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
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

func TestGetEditionGames(t *testing.T) {
	t.Parallel()

	type test struct {
		description        string
		editionID          openapi.EditionIDInPath
		gameVersions       []*service.GameVersionWithGame
		getEditionGamesErr error
		expectGames        []openapi.EditionGameResponse
		isErr              bool
		statusCode         int
	}

	now := time.Now()
	editionUUID := uuid.New()
	gameID := values.NewGameID()
	gameID2 := values.NewGameID()
	gameVersionID := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	imageID := values.NewGameImageID()
	videoID := values.NewGameVideoID()
	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID1UUID := uuid.UUID(fileID1)
	fileID2UUID := uuid.UUID(fileID2)
	game1 := domain.NewGame(
		gameID,
		values.NewGameName("テストゲーム1"),
		values.NewGameDescription("テスト説明1"),
		values.GameVisibilityTypePublic,
		now,
	)
	game2 := domain.NewGame(
		gameID2,
		values.NewGameName("テストゲーム2"),
		values.NewGameDescription("テスト説明2"),
		values.GameVisibilityTypePrivate,
		now,
	)
	gameLimited := domain.NewGame(
		gameID,
		values.NewGameName("テストゲーム"),
		values.NewGameDescription("テスト説明"),
		values.GameVisibilityTypeLimited,
		now,
	)
	gameVersion := domain.NewGameVersion(
		gameVersionID,
		values.NewGameVersionName("v1.0.0"),
		values.NewGameVersionDescription("リリース"),
		now,
	)
	gameVersion2 := domain.NewGameVersion(
		gameVersionID2,
		values.NewGameVersionName("v1.0.0"),
		values.NewGameVersionDescription("リリース"),
		now,
	)

	strURL := "https://example.com"
	questionnaireURL, err := url.Parse(strURL)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}
	urlValue := values.NewGameURLLink(questionnaireURL)

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			editionID:   editionUUID,
			gameVersions: []*service.GameVersionWithGame{
				{
					Game: game1,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							URL: option.NewOption(urlValue),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム1",
					Description: "テスト説明1",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Url:         &strURL,
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "ゲームURLとゲームファイルがnullでもエラーなし",
			editionID:   editionUUID,
			gameVersions: []*service.GameVersionWithGame{
				{
					Game: gameLimited,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets:      &service.Assets{},
						ImageID:     imageID,
						VideoID:     videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム",
					Description: "テスト説明",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "windowsでもエラーなし",
			editionID:   editionUUID,
			gameVersions: []*service.GameVersionWithGame{
				{
					Game: game1,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							Windows: option.NewOption(fileID1),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム1",
					Description: "テスト説明1",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Files: &openapi.GameVersionFiles{
							Win32: &fileID1UUID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "macでもエラーなし",
			editionID:   editionUUID,
			gameVersions: []*service.GameVersionWithGame{
				{
					Game: game1,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							Mac: option.NewOption(fileID1),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム1",
					Description: "テスト説明1",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Files: &openapi.GameVersionFiles{
							Darwin: &fileID1UUID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "jarでもエラーなし",
			editionID:   editionUUID,
			gameVersions: []*service.GameVersionWithGame{
				{
					Game: game1,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							Jar: option.NewOption(fileID1),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム1",
					Description: "テスト説明1",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Files: &openapi.GameVersionFiles{
							Jar: &fileID1UUID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "ファイルが複数あってももエラーなし",
			editionID:   editionUUID,
			gameVersions: []*service.GameVersionWithGame{
				{
					Game: game1,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							Jar:     option.NewOption(fileID1),
							Windows: option.NewOption(fileID2),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム1",
					Description: "テスト説明1",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Files: &openapi.GameVersionFiles{
							Jar:   &fileID1UUID,
							Win32: &fileID2UUID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "ファイルとurlが両方あってもエラーなし",
			editionID:   editionUUID,
			gameVersions: []*service.GameVersionWithGame{
				{
					Game: game1,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							Windows: option.NewOption(fileID1),
							URL:     option.NewOption(urlValue),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム1",
					Description: "テスト説明1",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Files: &openapi.GameVersionFiles{
							Win32: &fileID1UUID,
						},
						Url: &strURL,
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "2つ以上のゲームがリストに含まれていてもエラーなし",
			editionID:   editionUUID,
			gameVersions: []*service.GameVersionWithGame{
				{
					Game: game1,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							URL: option.NewOption(urlValue),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
				{
					Game: game2,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion2,
						Assets: &service.Assets{
							URL: option.NewOption(urlValue),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム1",
					Description: "テスト説明1",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Url:         &strURL,
					},
				},
				{
					Id:          uuid.UUID(gameID2),
					Name:        "テストゲーム2",
					Description: "テスト説明2",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID2),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Url:         &strURL,
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description:        "不正なeditionIDなので400",
			editionID:          editionUUID,
			getEditionGamesErr: service.ErrInvalidEditionID,
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		{
			description:        "サービス層でエラーが発生したので500",
			editionID:          editionUUID,
			getEditionGamesErr: errors.New("internal error"),
			isErr:              true,
			statusCode:         http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockEditionService := mock.NewMockEdition(ctrl)
			edition := NewEdition(mockEditionService)

			mockEditionService.
				EXPECT().
				GetEditionGameVersions(
					gomock.Any(),
					values.NewLauncherVersionIDFromUUID(testCase.editionID),
				).
				Return(testCase.gameVersions, testCase.getEditionGamesErr)

			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/api/v2/editions/%s/games", testCase.editionID), nil)

			err := edition.GetEditionGames(c, testCase.editionID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if assert.ErrorAs(t, err, &httpErr, "error should be *echo.HTTPError") {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					}
				} else {
					assert.Error(t, err)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)

			var res []openapi.EditionGameResponse
			err = json.NewDecoder(rec.Body).Decode(&res)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Len(t, res, len(testCase.expectGames))
			for i, game := range res {
				assert.Equal(t, testCase.expectGames[i].Id, game.Id)
				assert.Equal(t, testCase.expectGames[i].Name, game.Name)
				assert.Equal(t, testCase.expectGames[i].Description, game.Description)
				assert.WithinDuration(t, testCase.expectGames[i].CreatedAt, game.CreatedAt, 2*time.Second)

				assert.Equal(t, testCase.expectGames[i].Version.Id, game.Version.Id)
				assert.Equal(t, testCase.expectGames[i].Version.Name, game.Version.Name)
				assert.Equal(t, testCase.expectGames[i].Version.Description, game.Version.Description)
				assert.WithinDuration(t, testCase.expectGames[i].Version.CreatedAt, game.Version.CreatedAt, 2*time.Second)
				assert.Equal(t, testCase.expectGames[i].Version.Url, game.Version.Url)
				assert.Equal(t, testCase.expectGames[i].Version.Files, game.Version.Files)
				assert.Equal(t, testCase.expectGames[i].Version.ImageID, game.Version.ImageID)
				assert.Equal(t, testCase.expectGames[i].Version.VideoID, game.Version.VideoID)
			}
		})
	}
}

func TestPatchEditionGame(t *testing.T) {
	t.Parallel()

	type test struct {
		description           string
		editionID             openapi.EditionIDInPath
		reqBody               *openapi.PatchEditionGameRequest
		invalidBody           bool
		executeUpdateMock     bool
		launcherVersionID     values.LauncherVersionID
		gameVersionIDs        []values.GameVersionID
		updateEditionGamesErr error
		resultGameVersions    []*service.GameVersionWithGame
		expectGames           []openapi.EditionGameResponse
		isErr                 bool
		statusCode            int
	}

	now := time.Now()
	editionUUID := uuid.New()
	gameID := values.NewGameID()
	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID1UUID := uuid.UUID(fileID1)
	fileID2UUID := uuid.UUID(fileID2)
	imageID := values.NewGameImageID()
	videoID := values.NewGameVideoID()
	game := domain.NewGame(
		gameID,
		values.NewGameName("テストゲーム"),
		values.NewGameDescription("テスト説明"),
		values.GameVisibilityTypePublic,
		now,
	)
	gameVersion := domain.NewGameVersion(
		gameVersionID1,
		values.NewGameVersionName("v1.0.0"),
		values.NewGameVersionDescription("リリース"),
		now,
	)

	strURL := "https://example.com"
	questionnaireURL, err := url.Parse(strURL)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}
	urlValue := values.NewGameURLLink(questionnaireURL)

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{
					uuid.UUID(gameVersionID1),
					uuid.UUID(gameVersionID2),
				},
			},
			executeUpdateMock: true,
			launcherVersionID: values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs: []values.GameVersionID{
				gameVersionID1,
				gameVersionID2,
			},
			resultGameVersions: []*service.GameVersionWithGame{
				{
					Game: game,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets:      &service.Assets{},
						ImageID:     imageID,
						VideoID:     videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム",
					Description: "テスト説明",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "空のゲームバージョン一覧でもエラーなし",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{},
			},
			executeUpdateMock:  true,
			launcherVersionID:  values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs:     []values.GameVersionID{},
			resultGameVersions: []*service.GameVersionWithGame{},
			expectGames:        []openapi.EditionGameResponse{},
			statusCode:         http.StatusOK,
		},
		{
			description: "不正なリクエストボディなので400",
			editionID:   editionUUID,
			invalidBody: true,
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description: "不正なゲームバージョンIDなので400",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{uuid.UUID(gameVersionID1)},
			},
			executeUpdateMock:     true,
			launcherVersionID:     values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs:        []values.GameVersionID{gameVersionID1},
			updateEditionGamesErr: service.ErrInvalidEditionID,
			isErr:                 true,
			statusCode:            http.StatusBadRequest,
		},
		{
			description: "ErrDuplicateGameVersionなので400",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{
					uuid.UUID(gameVersionID1),
					uuid.UUID(gameVersionID1),
				},
			},
			executeUpdateMock:     true,
			launcherVersionID:     values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs:        []values.GameVersionID{gameVersionID1, gameVersionID1},
			updateEditionGamesErr: service.ErrDuplicateGameVersion,
			isErr:                 true,
			statusCode:            http.StatusBadRequest,
		},
		{
			description: "ErrDuplicateGameなので400",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{uuid.UUID(gameVersionID1)},
			},
			executeUpdateMock:     true,
			launcherVersionID:     values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs:        []values.GameVersionID{gameVersionID1},
			updateEditionGamesErr: service.ErrDuplicateGame,
			isErr:                 true,
			statusCode:            http.StatusBadRequest,
		},
		{
			description: "サービス層でエラーが発生したので500",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{uuid.UUID(gameVersionID1)},
			},
			executeUpdateMock:     true,
			launcherVersionID:     values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs:        []values.GameVersionID{gameVersionID1},
			updateEditionGamesErr: errors.New("internal error"),
			isErr:                 true,
			statusCode:            http.StatusInternalServerError,
		},
		{
			description: "windowsでもエラーなし",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{
					uuid.UUID(gameVersionID1),
					uuid.UUID(gameVersionID2),
				},
			},
			executeUpdateMock: true,
			launcherVersionID: values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs: []values.GameVersionID{
				gameVersionID1,
				gameVersionID2,
			},
			resultGameVersions: []*service.GameVersionWithGame{
				{
					Game: game,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							Windows: option.NewOption(fileID1),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム",
					Description: "テスト説明",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Files: &openapi.GameVersionFiles{
							Win32: &fileID1UUID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "macファイルでもエラーなし",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{uuid.UUID(gameVersionID1)},
			},
			executeUpdateMock: true,
			launcherVersionID: values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs:    []values.GameVersionID{gameVersionID1},
			resultGameVersions: []*service.GameVersionWithGame{
				{
					Game: game,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							Mac: option.NewOption(fileID2),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム",
					Description: "テスト説明",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Files: &openapi.GameVersionFiles{
							Darwin: &fileID2UUID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "jarファイルでもエラーなし",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{uuid.UUID(gameVersionID1)},
			},
			executeUpdateMock: true,
			launcherVersionID: values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs:    []values.GameVersionID{gameVersionID1},
			resultGameVersions: []*service.GameVersionWithGame{
				{
					Game: game,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							Jar: option.NewOption(fileID1),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム",
					Description: "テスト説明",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Files: &openapi.GameVersionFiles{
							Jar: &fileID1UUID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "ファイルが複数でもエラーなし",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{uuid.UUID(gameVersionID1)},
			},
			executeUpdateMock: true,
			launcherVersionID: values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs:    []values.GameVersionID{gameVersionID1},
			resultGameVersions: []*service.GameVersionWithGame{
				{
					Game: game,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							Windows: option.NewOption(fileID1),
							Mac:     option.NewOption(fileID2),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム",
					Description: "テスト説明",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Files: &openapi.GameVersionFiles{
							Win32:  &fileID1UUID,
							Darwin: &fileID2UUID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			description: "urlとファイルでもエラーなし",
			editionID:   editionUUID,
			reqBody: &openapi.PatchEditionGameRequest{
				GameVersionIDs: []uuid.UUID{uuid.UUID(gameVersionID1)},
			},
			executeUpdateMock: true,
			launcherVersionID: values.NewLauncherVersionIDFromUUID(editionUUID),
			gameVersionIDs:    []values.GameVersionID{gameVersionID1},
			resultGameVersions: []*service.GameVersionWithGame{
				{
					Game: game,
					GameVersion: service.GameVersionInfo{
						GameVersion: gameVersion,
						Assets: &service.Assets{
							URL:     option.NewOption(urlValue),
							Windows: option.NewOption(fileID1),
						},
						ImageID: imageID,
						VideoID: videoID,
					},
				},
			},
			expectGames: []openapi.EditionGameResponse{
				{
					Id:          uuid.UUID(gameID),
					Name:        "テストゲーム",
					Description: "テスト説明",
					CreatedAt:   now,
					Version: openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						ImageID:     uuid.UUID(imageID),
						VideoID:     uuid.UUID(videoID),
						Url:         &strURL,
						Files: &openapi.GameVersionFiles{
							Win32: &fileID1UUID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockEditionService := mock.NewMockEdition(ctrl)
			edition := NewEdition(mockEditionService)

			var bodyOpt bodyOpt
			if !testCase.invalidBody {
				bodyOpt = withJSONBody(t, testCase.reqBody)
			} else {
				bodyOpt = withStringBody(t, "invalid json")
			}

			c, _, rec := setupTestRequest(t, http.MethodPatch, fmt.Sprintf("/api/v2/editions/%s/games", testCase.editionID), bodyOpt)

			if testCase.executeUpdateMock {
				mockEditionService.
					EXPECT().
					UpdateEditionGameVersions(
						gomock.Any(),
						testCase.launcherVersionID,
						testCase.gameVersionIDs,
					).
					Return(testCase.resultGameVersions, testCase.updateEditionGamesErr)
			}

			err := edition.PatchEditionGame(c, testCase.editionID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if assert.ErrorAs(t, err, &httpErr, "error should be *echo.HTTPError") {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					}
				} else {
					assert.Error(t, err)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)

			if testCase.expectGames != nil {
				var res []openapi.EditionGameResponse
				err := json.NewDecoder(rec.Body).Decode(&res)
				assert.NoError(t, err)

				assert.Len(t, res, len(testCase.expectGames))
				for i, game := range res {
					assert.Equal(t, testCase.expectGames[i].Id, game.Id)
					assert.Equal(t, testCase.expectGames[i].Name, game.Name)
					assert.Equal(t, testCase.expectGames[i].Description, game.Description)
					assert.WithinDuration(t, testCase.expectGames[i].CreatedAt, game.CreatedAt, 2*time.Second)

					assert.Equal(t, testCase.expectGames[i].Version.Id, game.Version.Id)
					assert.Equal(t, testCase.expectGames[i].Version.Name, game.Version.Name)
					assert.Equal(t, testCase.expectGames[i].Version.Description, game.Version.Description)
					assert.WithinDuration(t, testCase.expectGames[i].Version.CreatedAt, game.Version.CreatedAt, 2*time.Second)
					assert.Equal(t, testCase.expectGames[i].Version.Url, game.Version.Url)
					assert.Equal(t, testCase.expectGames[i].Version.Files, game.Version.Files)
					assert.Equal(t, testCase.expectGames[i].Version.ImageID, game.Version.ImageID)
					assert.Equal(t, testCase.expectGames[i].Version.VideoID, game.Version.VideoID)
				}
			}
		})
	}
}
