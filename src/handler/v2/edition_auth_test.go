package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestGetProductKeys(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	editionID := uuid.New()

	activeProductKey1 := domain.NewProductKey(
		values.NewLauncherUserID(),
		values.NewLauncherUserProductKeyFromString("key"),
		values.LauncherUserStatusActive,
		time.Now(),
	)
	inactiveProductKey1 := domain.NewProductKey(
		values.NewLauncherUserID(),
		values.NewLauncherUserProductKeyFromString("inactive"),
		values.LauncherUserStatusInactive,
		time.Now(),
	)

	openapiActiveProductKey1 := openapi.ProductKey{
		Id:        openapi.ProductKeyID(activeProductKey1.GetID()),
		Key:       openapi.ProductKeyValue(activeProductKey1.GetProductKey()),
		Status:    openapi.Active,
		CreatedAt: activeProductKey1.GetCreatedAt(),
	}

	openapiInactiveProductKey1 := openapi.ProductKey{
		Id:        openapi.ProductKeyID(inactiveProductKey1.GetID()),
		Key:       openapi.ProductKeyValue(inactiveProductKey1.GetProductKey()),
		Status:    openapi.Revoked,
		CreatedAt: inactiveProductKey1.GetCreatedAt(),
	}

	active := openapi.Active
	revoked := openapi.Revoked
	invalid := openapi.ProductKeyStatus("a")

	testCases := map[string]struct {
		editionID             openapi.EditionIDInPath
		params                openapi.GetProductKeysParams
		executeGetProductKeys bool
		productKeys           []*domain.LauncherUser
		GetProductKeysErr     error
		resProductKeys        []openapi.ProductKey
		isErr                 bool
		err                   error
		statusCode            int
	}{
		"特に問題なし": {
			editionID:             editionID,
			executeGetProductKeys: true,
			productKeys:           []*domain.LauncherUser{activeProductKey1},
			resProductKeys:        []openapi.ProductKey{openapiActiveProductKey1},
		},
		"複数のプロダクトキーでも問題なし": {
			editionID:             editionID,
			executeGetProductKeys: true,
			productKeys:           []*domain.LauncherUser{activeProductKey1, inactiveProductKey1},
			resProductKeys:        []openapi.ProductKey{openapiActiveProductKey1, openapiInactiveProductKey1},
		},
		"プロダクトキーが無くても問題なし": {
			editionID:             editionID,
			executeGetProductKeys: true,
		},
		"statusがactiveでも問題なし": {
			editionID:             editionID,
			params:                openapi.GetProductKeysParams{Status: &active},
			executeGetProductKeys: true,
			productKeys:           []*domain.LauncherUser{activeProductKey1},
			resProductKeys:        []openapi.ProductKey{openapiActiveProductKey1},
		},
		"statusがinactiveでも問題なし": {
			editionID:             editionID,
			params:                openapi.GetProductKeysParams{Status: &revoked},
			executeGetProductKeys: true,
			productKeys:           []*domain.LauncherUser{inactiveProductKey1},
			resProductKeys:        []openapi.ProductKey{openapiInactiveProductKey1},
		},
		"statusが無効な値なので400": {
			editionID:  editionID,
			params:     openapi.GetProductKeysParams{Status: &invalid},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		"GetProductKeysがErrInvalidEditionIDなので400": {
			editionID:             editionID,
			executeGetProductKeys: true,
			GetProductKeysErr:     service.ErrInvalidEditionID,
			isErr:                 true,
			statusCode:            http.StatusBadRequest,
		},
		"GetProductKeysがエラーなので500": {
			editionID:             editionID,
			executeGetProductKeys: true,
			GetProductKeysErr:     errors.New("error"),
			isErr:                 true,
			statusCode:            http.StatusInternalServerError,
		},
		"GetProductKeysに無効なstatusがあったらそれを飛ばす": {
			editionID:             editionID,
			executeGetProductKeys: true,
			productKeys: []*domain.LauncherUser{activeProductKey1, domain.NewProductKey(
				values.NewLauncherUserID(),
				values.LauncherUserProductKey("key"),
				values.LauncherUserStatus(100),
				time.Now(),
			)},
			resProductKeys: []openapi.ProductKey{openapiActiveProductKey1},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockEditionAuthService := mock.NewMockEditionAuth(ctrl)
			editionAuth := NewEditionAuth(NewContext(), mockEditionAuthService)

			if testCase.executeGetProductKeys {
				var status types.Option[values.LauncherUserStatus]
				if testCase.params.Status != nil {
					switch *testCase.params.Status {
					case openapi.Active:
						status = types.NewOption(values.LauncherUserStatusActive)
					case openapi.Revoked:
						status = types.NewOption(values.LauncherUserStatusInactive)
					default:
						t.Fatalf("invalid params: %+v", testCase.params)
					}
				}
				mockEditionAuthService.
					EXPECT().
					GetProductKeys(gomock.Any(), values.NewLauncherVersionIDFromUUID(testCase.editionID), service.GetProductKeysParams{Status: status}).
					Return(testCase.productKeys, testCase.GetProductKeysErr)
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v2/editions/%s/keys", testCase.editionID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := editionAuth.GetProductKeys(c, testCase.editionID, testCase.params)

			if testCase.isErr {
				if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
				} else {
					assert.Error(t, err)
				}

				var httpErr *echo.HTTPError
				assert.ErrorAs(t, err, &httpErr)
				assert.Equal(t, testCase.statusCode, httpErr.Code)
			} else {
				assert.NoError(t, err)
			}

			if err != nil {
				return
			}

			var resProductKeys []openapi.ProductKey
			err = json.NewDecoder(rec.Body).Decode(&resProductKeys)
			require.NoError(t, err)

			assert.Len(t, resProductKeys, len(testCase.resProductKeys))
			for i, productKey := range resProductKeys {
				expectedProductKey := testCase.resProductKeys[i]
				assert.Equal(t, expectedProductKey.Id, productKey.Id)
				assert.Equal(t, expectedProductKey.Key, productKey.Key)
				assert.Equal(t, expectedProductKey.Status, productKey.Status)
				assert.WithinDuration(t, expectedProductKey.CreatedAt, productKey.CreatedAt, 0)
			}
		})
	}
}
