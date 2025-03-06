package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestPostProductKey(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	editionID := uuid.New()
	num := 5

	productKey1 := domain.NewProductKey(
		values.NewLauncherUserID(),
		values.NewLauncherUserProductKeyFromString("key1"),
		values.LauncherUserStatusActive,
		time.Now(),
	)
	productKey2 := domain.NewProductKey(
		values.NewLauncherUserID(),
		values.NewLauncherUserProductKeyFromString("key2"),
		values.LauncherUserStatusActive,
		time.Now(),
	)

	openapiProductKey1 := openapi.ProductKey{
		Id:        openapi.ProductKeyID(productKey1.GetID()),
		Key:       openapi.ProductKeyValue(productKey1.GetProductKey()),
		Status:    openapi.Active,
		CreatedAt: productKey1.GetCreatedAt(),
	}

	openapiProductKey2 := openapi.ProductKey{
		Id:        openapi.ProductKeyID(productKey2.GetID()),
		Key:       openapi.ProductKeyValue(productKey2.GetProductKey()),
		Status:    openapi.Active,
		CreatedAt: productKey2.GetCreatedAt(),
	}

	testCases := map[string]struct {
		editionID             openapi.EditionIDInPath
		params                openapi.PostProductKeyParams
		executeGetProductKeys bool
		productKeys           []*domain.LauncherUser
		GenerateProductKeyErr error
		resProductKeys        []openapi.ProductKey
		isErr                 bool
		err                   error
		statusCode            int
	}{
		"特に問題なし": {
			editionID:             editionID,
			params:                openapi.PostProductKeyParams{Num: num},
			executeGetProductKeys: true,
			productKeys:           []*domain.LauncherUser{productKey1, productKey2},
			resProductKeys:        []openapi.ProductKey{openapiProductKey1, openapiProductKey2},
		},
		"ErrInvalidEditionIDなので400": {
			editionID:             editionID,
			params:                openapi.PostProductKeyParams{Num: num},
			executeGetProductKeys: true,
			GenerateProductKeyErr: service.ErrInvalidEditionID,
			isErr:                 true,
			statusCode:            http.StatusBadRequest,
		},
		"ErrInvalidKeyNumなので400": {
			editionID:             editionID,
			params:                openapi.PostProductKeyParams{Num: num},
			executeGetProductKeys: true,
			GenerateProductKeyErr: service.ErrInvalidKeyNum,
			isErr:                 true,
			statusCode:            http.StatusBadRequest,
		},
		"エラーが発生して500": {
			editionID:             editionID,
			params:                openapi.PostProductKeyParams{Num: num},
			executeGetProductKeys: true,
			GenerateProductKeyErr: errors.New("error"),
			isErr:                 true,
			statusCode:            http.StatusInternalServerError,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockEditionAuthService := mock.NewMockEditionAuth(ctrl)
			editionAuth := NewEditionAuth(NewContext(), mockEditionAuthService)

			if testCase.executeGetProductKeys {
				mockEditionAuthService.
					EXPECT().
					GenerateProductKey(gomock.Any(), values.NewLauncherVersionIDFromUUID(testCase.editionID), uint(testCase.params.Num)).
					Return(testCase.productKeys, testCase.GenerateProductKeyErr)
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v2/editions/%s/keys", testCase.editionID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := editionAuth.PostProductKey(c, testCase.editionID, testCase.params)

			if testCase.isErr {
				if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
				} else {
					assert.Error(t, err)
				}

				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					assert.ErrorAs(t, err, &httpErr)
					assert.Equal(t, testCase.statusCode, httpErr.Code)
				}
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
func TestPostActivateProductKey(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	productKeyID := uuid.New()

	activeProductKey := domain.NewProductKey(
		values.NewLauncherUserIDFromUUID(productKeyID),
		values.NewLauncherUserProductKeyFromString("key"),
		values.LauncherUserStatusActive,
		time.Now(),
	)

	openapiActiveProductKey := openapi.ProductKey{
		Id:        openapi.ProductKeyID(activeProductKey.GetID()),
		Key:       openapi.ProductKeyValue(activeProductKey.GetProductKey()),
		Status:    openapi.Active,
		CreatedAt: activeProductKey.GetCreatedAt(),
	}

	testCases := map[string]struct {
		productKeyID              openapi.ProductKeyIDInPath
		executeActivateProductKey bool
		productKey                *domain.LauncherUser
		ActivateProductKeyErr     error
		resProductKey             openapi.ProductKey
		isErr                     bool
		err                       error
		statusCode                int
	}{
		"特に問題なし": {
			productKeyID:              productKeyID,
			executeActivateProductKey: true,
			productKey:                activeProductKey,
			resProductKey:             openapiActiveProductKey,
		},
		"ErrInvalidProductKeyなので400": {
			productKeyID:              productKeyID,
			executeActivateProductKey: true,
			ActivateProductKeyErr:     service.ErrInvalidProductKey,
			isErr:                     true,
			statusCode:                http.StatusBadRequest,
		},
		"ErrKeyAlreadyActivatedなので404": {
			productKeyID:              productKeyID,
			executeActivateProductKey: true,
			ActivateProductKeyErr:     service.ErrKeyAlreadyActivated,
			isErr:                     true,
			statusCode:                http.StatusNotFound,
		},
		"エラーが発生して500": {
			productKeyID:              productKeyID,
			executeActivateProductKey: true,
			ActivateProductKeyErr:     errors.New("error"),
			isErr:                     true,
			statusCode:                http.StatusInternalServerError,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockEditionAuthService := mock.NewMockEditionAuth(ctrl)
			editionAuth := NewEditionAuth(NewContext(), mockEditionAuthService)

			if testCase.executeActivateProductKey {
				mockEditionAuthService.
					EXPECT().
					ActivateProductKey(gomock.Any(), values.NewLauncherUserIDFromUUID(testCase.productKeyID)).
					Return(testCase.productKey, testCase.ActivateProductKeyErr)
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v2/editions/keys/%s/activate", testCase.productKeyID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := editionAuth.PostActivateProductKey(c, openapi.EditionIDInPath{}, testCase.productKeyID)

			if testCase.isErr {
				if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
				} else {
					assert.Error(t, err)
				}

				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					assert.ErrorAs(t, err, &httpErr)
					assert.Equal(t, testCase.statusCode, httpErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}

			if err != nil {
				return
			}

			var resProductKey openapi.ProductKey
			err = json.NewDecoder(rec.Body).Decode(&resProductKey)
			require.NoError(t, err)

			assert.Equal(t, testCase.resProductKey.Id, resProductKey.Id)
			assert.Equal(t, testCase.resProductKey.Key, resProductKey.Key)
			assert.Equal(t, testCase.resProductKey.Status, resProductKey.Status)
			assert.WithinDuration(t, testCase.resProductKey.CreatedAt, resProductKey.CreatedAt, 0)
		})
	}
}
func TestPostRevokeProductKey(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	productKeyID := uuid.New()

	revokedProductKey := domain.NewProductKey(
		values.NewLauncherUserIDFromUUID(productKeyID),
		values.NewLauncherUserProductKeyFromString("key"),
		values.LauncherUserStatusInactive,
		time.Now(),
	)

	openapiRevokedProductKey := openapi.ProductKey{
		Id:        openapi.ProductKeyID(revokedProductKey.GetID()),
		Key:       openapi.ProductKeyValue(revokedProductKey.GetProductKey()),
		Status:    openapi.Revoked,
		CreatedAt: revokedProductKey.GetCreatedAt(),
	}

	testCases := map[string]struct {
		productKeyID            openapi.ProductKeyIDInPath
		executeRevokeProductKey bool
		productKey              *domain.LauncherUser
		RevokeProductKeyErr     error
		resProductKey           openapi.ProductKey
		isErr                   bool
		err                     error
		statusCode              int
	}{
		"特に問題なし": {
			productKeyID:            productKeyID,
			executeRevokeProductKey: true,
			productKey:              revokedProductKey,
			resProductKey:           openapiRevokedProductKey,
		},
		"ErrInvalidProductKeyなので400": {
			productKeyID:            productKeyID,
			executeRevokeProductKey: true,
			RevokeProductKeyErr:     service.ErrInvalidProductKey,
			isErr:                   true,
			statusCode:              http.StatusBadRequest,
		},
		"ErrKeyAlreadyRevokedなので404": {
			productKeyID:            productKeyID,
			executeRevokeProductKey: true,
			RevokeProductKeyErr:     service.ErrKeyAlreadyRevoked,
			isErr:                   true,
			statusCode:              http.StatusNotFound,
		},
		"エラーが発生して500": {
			productKeyID:            productKeyID,
			executeRevokeProductKey: true,
			RevokeProductKeyErr:     errors.New("error"),
			isErr:                   true,
			statusCode:              http.StatusInternalServerError,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockEditionAuthService := mock.NewMockEditionAuth(ctrl)
			editionAuth := NewEditionAuth(NewContext(), mockEditionAuthService)

			if testCase.executeRevokeProductKey {
				mockEditionAuthService.
					EXPECT().
					RevokeProductKey(gomock.Any(), values.NewLauncherUserIDFromUUID(testCase.productKeyID)).
					Return(testCase.productKey, testCase.RevokeProductKeyErr)
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v2/editions/keys/%s/revoke", testCase.productKeyID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := editionAuth.PostRevokeProductKey(c, openapi.EditionIDInPath{}, testCase.productKeyID)

			if testCase.isErr {
				if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
				} else {
					assert.Error(t, err)
				}

				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					assert.ErrorAs(t, err, &httpErr)
					assert.Equal(t, testCase.statusCode, httpErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}

			if err != nil {
				return
			}

			var resProductKey openapi.ProductKey
			err = json.NewDecoder(rec.Body).Decode(&resProductKey)
			require.NoError(t, err)

			assert.Equal(t, testCase.resProductKey.Id, resProductKey.Id)
			assert.Equal(t, testCase.resProductKey.Key, resProductKey.Key)
			assert.Equal(t, testCase.resProductKey.Status, resProductKey.Status)
			assert.WithinDuration(t, testCase.resProductKey.CreatedAt, resProductKey.CreatedAt, 0)
		})
	}
}
func TestPostEditionAuthorize(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	validKey, err := values.NewLauncherUserProductKey()
	require.NoError(t, err)
	validKeyStr := string(validKey)
	invalidKeyFormat := "invalidKeyFormat"

	validAccessToken := domain.NewLauncherSession(
		values.NewLauncherSessionID(),
		values.NewLauncherSessionAccessTokenFromString("accessToken"),
		time.Now().Add(time.Hour),
	)

	validRequestBody := func(t *testing.T) io.Reader {
		body, err := json.Marshal(openapi.EditionAuthorizeRequest{Key: string(validKeyStr)})
		require.NoError(t, err)
		return bytes.NewBuffer(body)
	}

	testCases := map[string]struct {
		requestBody             func(t *testing.T) io.Reader
		executeAuthorizeEdition bool
		authorizeEditionKey     values.LauncherUserProductKey
		authorizeEditionToken   *domain.LauncherSession
		authorizeEditionErr     error
		isErr                   bool
		err                     error
		statusCode              int
	}{
		"特に問題なし": {
			requestBody:             validRequestBody,
			executeAuthorizeEdition: true,
			authorizeEditionKey:     values.NewLauncherUserProductKeyFromString(validKeyStr),
			authorizeEditionToken:   validAccessToken,
		},
		"リクエストボディが無効なので400": {
			requestBody: func(_ *testing.T) io.Reader {
				body := `{"invalid": "body"}`
				return strings.NewReader(body)
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		"プロダクトキーが無効な形式なので400": {
			requestBody: func(t *testing.T) io.Reader {
				body, err := json.Marshal(openapi.EditionAuthorizeRequest{Key: invalidKeyFormat})
				require.NoError(t, err)
				return bytes.NewBuffer(body)
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		"AuthorizeEditionがErrInvalidProductKeyなので400": {
			requestBody:             validRequestBody,
			executeAuthorizeEdition: true,
			authorizeEditionKey:     values.NewLauncherUserProductKeyFromString(validKeyStr),
			authorizeEditionErr:     service.ErrInvalidProductKey,
			isErr:                   true,
			statusCode:              http.StatusBadRequest,
		},
		"AuthorizeEditionがエラーなので500": {
			requestBody:             validRequestBody,
			executeAuthorizeEdition: true,
			authorizeEditionKey:     values.NewLauncherUserProductKeyFromString(validKeyStr),
			authorizeEditionErr:     errors.New("error"),
			isErr:                   true,
			statusCode:              http.StatusInternalServerError,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockEditionAuthService := mock.NewMockEditionAuth(ctrl)
			editionAuth := NewEditionAuth(NewContext(), mockEditionAuthService)

			if testCase.executeAuthorizeEdition {
				mockEditionAuthService.
					EXPECT().
					AuthorizeEdition(gomock.Any(), testCase.authorizeEditionKey).
					Return(testCase.authorizeEditionToken, testCase.authorizeEditionErr)
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v2/editions/authorize", testCase.requestBody(t))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := editionAuth.PostEditionAuthorize(c)

			if testCase.isErr {
				if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
				} else {
					assert.Error(t, err)
				}

				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					assert.ErrorAs(t, err, &httpErr)
					assert.Equal(t, testCase.statusCode, httpErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}

			if err != nil {
				return
			}

			var resAccessToken openapi.EditionAccessToken
			err = json.NewDecoder(rec.Body).Decode(&resAccessToken)
			require.NoError(t, err)

			assert.Equal(t, string(validAccessToken.GetAccessToken()), resAccessToken.AccessToken)
			assert.WithinDuration(t, validAccessToken.GetExpiresAt(), resAccessToken.ExpiresAt, 0)
		})
	}
}
