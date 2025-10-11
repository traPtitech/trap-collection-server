package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func TestPostGamePlayLogStart(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	editionID := values.NewLauncherVersionID()
	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()
	gameStartTime := time.Now()
	reqBody := openapi.PostGamePlayLogStartRequest{
		GameVersionID: openapi.GameVersionID(gameVersionID),
		EditionID:     openapi.EditionID(editionID),
		GameID:        openapi.GameID(gameID),
		StartTime:     gameStartTime,
	}

	playLogID := values.NewGamePlayLogID()

	testCases := map[string]struct {
		editionID            values.LauncherVersionID
		gameID               values.GameID
		invalidReqBody       bool
		reqBody              openapi.PostGamePlayLogStartRequest
		executeCreatePlayLog bool
		playLog              *domain.GamePlayLog
		CreatePlayLogErr     error
		isError              bool
		statusCode           int
		resBody              openapi.PostGamePlayLogStartResponse
	}{
		"request bodyが不正なのでエラー": {
			editionID:      editionID,
			gameID:         gameID,
			invalidReqBody: true,
			isError:        true,
			statusCode:     http.StatusBadRequest,
		},
		"CreatePlayLogがErrInvalidEditionなので404": {
			editionID:            editionID,
			gameID:               gameID,
			reqBody:              reqBody,
			executeCreatePlayLog: true,
			CreatePlayLogErr:     service.ErrInvalidEdition,
			isError:              true,
			statusCode:           http.StatusNotFound,
		},
		"CreatePlayLogがErrInvalidGameなので404": {
			editionID:            editionID,
			gameID:               gameID,
			reqBody:              reqBody,
			executeCreatePlayLog: true,
			CreatePlayLogErr:     service.ErrInvalidGame,
			isError:              true,
			statusCode:           http.StatusNotFound,
		},
		"CreatePlayLogがErrInvalidGameVersionなので404": {
			editionID:            editionID,
			gameID:               gameID,
			reqBody:              reqBody,
			executeCreatePlayLog: true,
			CreatePlayLogErr:     service.ErrInvalidGameVersion,
			isError:              true,
			statusCode:           http.StatusNotFound,
		},
		"CreatePlayLogがその他のエラーなので500": {
			editionID:            editionID,
			gameID:               gameID,
			reqBody:              reqBody,
			executeCreatePlayLog: true,
			CreatePlayLogErr:     assert.AnError,
			isError:              true,
			statusCode:           http.StatusInternalServerError,
		},
		"CreatePlayLogが成功するので201": {
			editionID:            editionID,
			gameID:               gameID,
			reqBody:              reqBody,
			executeCreatePlayLog: true,
			playLog:              domain.NewGamePlayLog(playLogID, editionID, gameID, gameVersionID, gameStartTime, nil, time.Now(), time.Now()),
			statusCode:           http.StatusCreated,
			resBody: openapi.PostGamePlayLogStartResponse{
				PlayLogID: openapi.GamePlayLogID(playLogID),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			serviceMock := mock.NewMockGamePlayLogV2(ctrl)
			h := NewGamePlayLog(serviceMock)

			gameVersionID := values.NewGameVersionIDFromUUID(testCase.reqBody.GameVersionID)

			if testCase.executeCreatePlayLog {
				serviceMock.
					EXPECT().
					CreatePlayLog(
						gomock.Any(),
						testCase.editionID,
						testCase.gameID,
						gameVersionID,
						gomock.Cond(func(startTime time.Time) bool { return startTime.Sub(testCase.reqBody.StartTime).Abs() < time.Second }), // JSONのエンコードとデコードで精度がずれるため
					).
					Return(testCase.playLog, testCase.CreatePlayLogErr)
			}

			var body bodyOpt
			if testCase.invalidReqBody {
				body = withStringBody(t, "invalid")
			} else {
				body = withJSONBody(t, testCase.reqBody)
			}

			url := fmt.Sprintf("/editions/%s/games/%s/plays/start",
				uuid.UUID(testCase.editionID).String(), uuid.UUID(testCase.gameID).String())
			c, _, rec := setupTestRequest(t, http.MethodPost, url, body)

			err := h.PostGamePlayLogStart(c, openapi.EditionIDInPath(testCase.editionID), openapi.GameIDInPath(testCase.gameID))

			if testCase.isError {
				var httpError *echo.HTTPError
				assert.ErrorAs(t, err, &httpError)
				assert.Equal(t, testCase.statusCode, httpError.Code)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)

			var resBody openapi.PostGamePlayLogStartResponse
			assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resBody))
			assert.Equal(t, testCase.resBody.PlayLogID, resBody.PlayLogID)
		})
	}

}
