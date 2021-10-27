package v1

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestPostKeyGenerate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)

	launcherAuthHandler := NewLauncherAuth(mockLauncherAuthService)

	type test struct {
		description               string
		request                   openapi.ProductKeyGen
		executeCreateLauncherUser bool
		versionID                 values.LauncherVersionID
		launcherUsers             []*domain.LauncherUser
		CreateLauncherUserErr     error
		expect                    []*openapi.ProductKey
		isErr                     bool
		err                       error
		statusCode                int
	}

	versionID := values.NewLauncherVersionID()

	productKey1, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	productKey2, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	testCases := []test{
		{
			description: "エラーなしなので問題なし",
			request: openapi.ProductKeyGen{
				Num:     1,
				Version: uuid.UUID(versionID).String(),
			},
			executeCreateLauncherUser: true,
			versionID:                 versionID,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey1,
				),
			},
			expect: []*openapi.ProductKey{
				{
					Key: string(productKey1),
				},
			},
		},
		{
			description: "numが0なので400",
			request: openapi.ProductKeyGen{
				Num:     0,
				Version: uuid.UUID(versionID).String(),
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "numが負なので400",
			request: openapi.ProductKeyGen{
				Num:     -1,
				Version: uuid.UUID(versionID).String(),
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "versionIDがuuidでないので400",
			request: openapi.ProductKeyGen{
				Num:     1,
				Version: "1.0.0",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "versionIDが空文字なので400",
			request: openapi.ProductKeyGen{
				Num:     1,
				Version: "",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "CreateLauncherUserがErrInvalidLauncherVersionなので400",
			request: openapi.ProductKeyGen{
				Num:     1,
				Version: uuid.UUID(versionID).String(),
			},
			executeCreateLauncherUser: true,
			versionID:                 versionID,
			CreateLauncherUserErr:     service.ErrInvalidLauncherVersion,
			isErr:                     true,
			statusCode:                http.StatusBadRequest,
		},
		{
			description: "CreateLauncherUserがエラー(ErrInvalidLauncherVersion以外)なので500",
			request: openapi.ProductKeyGen{
				Num:     1,
				Version: uuid.UUID(versionID).String(),
			},
			executeCreateLauncherUser: true,
			versionID:                 versionID,
			CreateLauncherUserErr:     errors.New("error"),
			isErr:                     true,
			statusCode:                http.StatusInternalServerError,
		},
		{
			description: "launcherUserが複数でも問題なし",
			request: openapi.ProductKeyGen{
				Num:     2,
				Version: uuid.UUID(versionID).String(),
			},
			executeCreateLauncherUser: true,
			versionID:                 versionID,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey1,
				),
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey2,
				),
			},
			expect: []*openapi.ProductKey{
				{
					Key: string(productKey1),
				},
				{
					Key: string(productKey2),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeCreateLauncherUser {
				mockLauncherAuthService.
					EXPECT().
					CreateLauncherUser(ctx, testCase.versionID, int(testCase.request.Num)).
					Return(testCase.launcherUsers, testCase.CreateLauncherUserErr)
			}

			productKeys, err := launcherAuthHandler.PostKeyGenerate(&testCase.request)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, len(testCase.expect), len(productKeys))
			for i, expect := range testCase.expect {
				assert.Equal(t, *expect, *productKeys[i])
			}
		})
	}
}

func TestPostLauncherLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)

	launcherAuthHandler := NewLauncherAuth(mockLauncherAuthService)

	type test struct {
		description          string
		request              openapi.ProductKey
		executeLoginLauncher bool
		productKey           values.LauncherUserProductKey
		launcherSession      *domain.LauncherSession
		LoginLauncherErr     error
		expect               openapi.LauncherAuthToken
		isErr                bool
		err                  error
		statusCode           int
	}

	accessToken, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	now := time.Now()

	testCases := []test{
		{
			description: "エラーなしなので問題なし",
			request: openapi.ProductKey{
				Key: "abcde-fghij-klmno-pqrst-uvwxy",
			},
			executeLoginLauncher: true,
			productKey:           values.LauncherUserProductKey("abcde-fghij-klmno-pqrst-uvwxy"),
			launcherSession: domain.NewLauncherSession(
				values.NewLauncherSessionID(),
				accessToken,
				now.Add(time.Hour),
			),
			expect: openapi.LauncherAuthToken{
				AccessToken: string(accessToken),
				ExpiresIn:   int32(time.Until(now.Add(time.Hour)).Seconds()),
			},
		},
		{
			description: "productKeyが誤った形式なので400",
			request: openapi.ProductKey{
				Key: "abcde",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "productKeyが空文字なので400",
			request: openapi.ProductKey{
				Key: "",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "LoginLauncherがErrInvalidLauncherProductKeyなので400",
			request: openapi.ProductKey{
				Key: "abcde-fghij-klmno-pqrst-uvwxy",
			},
			executeLoginLauncher: true,
			productKey:           values.LauncherUserProductKey("abcde-fghij-klmno-pqrst-uvwxy"),
			LoginLauncherErr:     service.ErrInvalidLauncherUserProductKey,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description: "LoginLauncherがエラー(ErrInvalidLauncherProductKey以外)なので500",
			request: openapi.ProductKey{
				Key: "abcde-fghij-klmno-pqrst-uvwxy",
			},
			executeLoginLauncher: true,
			productKey:           values.LauncherUserProductKey("abcde-fghij-klmno-pqrst-uvwxy"),
			LoginLauncherErr:     errors.New("error"),
			isErr:                true,
			statusCode:           http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeLoginLauncher {
				mockLauncherAuthService.
					EXPECT().
					LoginLauncher(ctx, testCase.productKey).
					Return(testCase.launcherSession, testCase.LoginLauncherErr)
			}

			token, err := launcherAuthHandler.PostLauncherLogin(&testCase.request)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, testCase.expect, *token)
		})
	}
}

func TestDeleteProductKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)

	launcherAuthHandler := NewLauncherAuth(mockLauncherAuthService)

	type test struct {
		description             string
		requestProductKeyID     string
		executeRevokeProductKey bool
		launcherUserID          values.LauncherUserID
		RevokeProductKeyErr     error
		isErr                   bool
		err                     error
		statusCode              int
	}

	launcherUserID := values.NewLauncherUserID()

	testCases := []test{
		{
			description:             "エラーなしなので問題なし",
			requestProductKeyID:     uuid.UUID(launcherUserID).String(),
			executeRevokeProductKey: true,
			launcherUserID:          launcherUserID,
		},
		{
			description:         "productKeyIDがuuidでないので400",
			requestProductKeyID: "abcde",
			isErr:               true,
			statusCode:          http.StatusBadRequest,
		},
		{
			description:         "productKeyIDが空文字なので400",
			requestProductKeyID: "",
			isErr:               true,
			statusCode:          http.StatusBadRequest,
		},
		{
			description:             "RevokeProductKeyがエラー(ErrInvalidLauncherProductKey以外)なので500",
			requestProductKeyID:     uuid.UUID(launcherUserID).String(),
			executeRevokeProductKey: true,
			launcherUserID:          launcherUserID,
			RevokeProductKeyErr:     errors.New("error"),
			isErr:                   true,
			statusCode:              http.StatusInternalServerError,
		},
		{
			description:             "RevokeProductKeyがErrInvalidLauncherUserなので400",
			requestProductKeyID:     uuid.UUID(launcherUserID).String(),
			executeRevokeProductKey: true,
			launcherUserID:          launcherUserID,
			RevokeProductKeyErr:     service.ErrInvalidLauncherUser,
			isErr:                   true,
			statusCode:              http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeRevokeProductKey {
				mockLauncherAuthService.
					EXPECT().
					RevokeProductKey(ctx, testCase.launcherUserID).
					Return(testCase.RevokeProductKeyErr)
			}

			err := launcherAuthHandler.DeleteProductKey(testCase.requestProductKeyID)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetProductKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)

	launcherAuthHandler := NewLauncherAuth(mockLauncherAuthService)

	type test struct {
		description              string
		requestLauncherVersionID string
		executeGetLauncherUsers  bool
		launcherVersionID        values.LauncherVersionID
		launcherUsers            []*domain.LauncherUser
		GetLauncherUsersErr      error
		expect                   []*openapi.ProductKeyDetail
		isErr                    bool
		err                      error
		statusCode               int
	}

	launcherVersionID := values.NewLauncherVersionID()
	launcherUserID1 := values.NewLauncherUserID()
	launcherUserID2 := values.NewLauncherUserID()

	productKey1, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	productKey2, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	testCases := []test{
		{
			description:              "エラーなしなので問題なし",
			requestLauncherVersionID: uuid.UUID(launcherVersionID).String(),
			executeGetLauncherUsers:  true,
			launcherVersionID:        launcherVersionID,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					launcherUserID1,
					productKey1,
				),
			},
			expect: []*openapi.ProductKeyDetail{
				{
					Id:  uuid.UUID(launcherUserID1).String(),
					Key: string(productKey1),
				},
			},
		},
		{
			description:              "ユーザーが複数でも問題なし",
			requestLauncherVersionID: uuid.UUID(launcherVersionID).String(),
			executeGetLauncherUsers:  true,
			launcherVersionID:        launcherVersionID,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					launcherUserID1,
					productKey1,
				),
				domain.NewLauncherUser(
					launcherUserID2,
					productKey2,
				),
			},
			expect: []*openapi.ProductKeyDetail{
				{
					Id:  uuid.UUID(launcherUserID1).String(),
					Key: string(productKey1),
				},
				{
					Id:  uuid.UUID(launcherUserID2).String(),
					Key: string(productKey2),
				},
			},
		},
		{
			description:              "ユーザーがいなくても問題なし",
			requestLauncherVersionID: uuid.UUID(launcherVersionID).String(),
			executeGetLauncherUsers:  true,
			launcherVersionID:        launcherVersionID,
			launcherUsers:            []*domain.LauncherUser{},
			expect:                   []*openapi.ProductKeyDetail{},
		},
		{
			description:              "launcherVersionIDがuuidでないので400",
			requestLauncherVersionID: "abcde",
			isErr:                    true,
			statusCode:               http.StatusBadRequest,
		},
		{
			description:              "launcherVersionIDが空文字なので400",
			requestLauncherVersionID: "",
			isErr:                    true,
			statusCode:               http.StatusBadRequest,
		},
		{
			description:              "GetLauncherUsersがエラー(ErrInvalidLauncherVersion以外)なので500",
			requestLauncherVersionID: uuid.UUID(launcherVersionID).String(),
			executeGetLauncherUsers:  true,
			launcherVersionID:        launcherVersionID,
			GetLauncherUsersErr:      errors.New("error"),
			isErr:                    true,
			statusCode:               http.StatusInternalServerError,
		},
		{
			description:              "GetLauncherUsersがErrInvalidLauncherVersionなので400",
			requestLauncherVersionID: uuid.UUID(launcherVersionID).String(),
			executeGetLauncherUsers:  true,
			launcherVersionID:        launcherVersionID,
			GetLauncherUsersErr:      service.ErrInvalidLauncherVersion,
			isErr:                    true,
			statusCode:               http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeGetLauncherUsers {
				mockLauncherAuthService.
					EXPECT().
					GetLauncherUsers(ctx, testCase.launcherVersionID).
					Return(testCase.launcherUsers, testCase.GetLauncherUsersErr)
			}

			actualProductKeys, err := launcherAuthHandler.GetProductKeys(testCase.requestLauncherVersionID)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, len(testCase.expect), len(actualProductKeys))
			for i, expect := range testCase.expect {
				assert.Equal(t, expect.Id, actualProductKeys[i].Id)
				assert.Equal(t, expect.Key, actualProductKeys[i].Key)
			}
		})
	}
}
