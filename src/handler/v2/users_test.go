package v2

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
	"github.com/traPtitech/trap-collection-server/src/handler/v1/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestGetMe(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock.NewMockUser(ctrl)
	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := common.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	userHandler := NewUser(session, mockUserService)

	type test struct {
		description      string
		sessionExist     bool
		authSessionExist bool
		accessToken      string
		expiresAt        time.Time
		executeGetMe     bool
		userInfo         *service.UserInfo
		GetMeErr         error
		user             *openapi.User
		isErr            bool
		err              error
		statusCode       int
	}

	id1 := uuid.New()

	testCases := []test{
		{
			description:      "特に問題ないのでエラーなし",
			sessionExist:     true,
			authSessionExist: true,
			accessToken:      "accessToken",
			expiresAt:        time.Now(),
			executeGetMe:     true,
			userInfo: service.NewUserInfo(
				values.NewTrapMemberID(id1),
				"mazrean",
				values.TrapMemberStatusActive,
			),
			user: &openapi.User{
				Id:   id1.String(),
				Name: "mazrean",
			},
		},
		{
			// 実際にはmiddlewareで弾かれるが、念の為確認
			description:  "sessionが存在しないのでauthSessionも存在せず500",
			sessionExist: false,
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			// 実際にはmiddlewareで弾かれるが、念の為確認
			description:      "authSessionが存在しないので500",
			sessionExist:     true,
			authSessionExist: false,
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
		{
			description:      "GetMeがエラーなので500",
			sessionExist:     true,
			authSessionExist: true,
			accessToken:      "accessToken",
			expiresAt:        time.Now(),
			executeGetMe:     true,
			GetMeErr:         errors.New("error"),
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/users/me", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.New(req)
				if err != nil {
					t.Fatal(err)
				}

				if testCase.authSessionExist {
					sess.Values[accessTokenSessionKey] = testCase.accessToken
					sess.Values[expiresAtSessionKey] = testCase.expiresAt
				}

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)

				sess, err = session.Get(req)
				if err != nil {
					t.Fatal(err)
				}

				setCookieHeader(c)
			}

			if testCase.executeGetMe {
				mockUserService.
					EXPECT().
					GetMe(gomock.Any(), gomock.Any()).
					Return(testCase.userInfo, testCase.GetMeErr)
			}

			user := userHandler.GetMe(c)

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
			if err != nil {
				return
			}

			assert.Equal(t, *testCase.user, user)
		})
	}
}

func TestGetUsers(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock.NewMockUser(ctrl)
	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := common.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	userHandler := NewUser(session, mockUserService)

	type test struct {
		description             string
		sessionExist            bool
		authSessionExist        bool
		accessToken             string
		expiresAt               time.Time
		executeGetAllActiveUser bool
		userInfos               []*service.UserInfo
		GetAllActiveUserErr     error
		users                   []*openapi.User
		isErr                   bool
		err                     error
		statusCode              int
	}

	id1 := uuid.New()
	id2 := uuid.New()

	testCases := []test{
		{
			description:             "特に問題ないのでエラーなし",
			sessionExist:            true,
			authSessionExist:        true,
			accessToken:             "accessToken",
			expiresAt:               time.Now(),
			executeGetAllActiveUser: true,
			userInfos: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"mazrean",
					values.TrapMemberStatusActive,
				),
			},
			users: []*openapi.User{
				{
					Id:   id1.String(),
					Name: "mazrean",
				},
			},
		},
		{
			description:             "userが複数でもエラーなし",
			sessionExist:            true,
			authSessionExist:        true,
			accessToken:             "accessToken",
			expiresAt:               time.Now(),
			executeGetAllActiveUser: true,
			userInfos: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"mazrean",
					values.TrapMemberStatusActive,
				),
				service.NewUserInfo(
					values.NewTrapMemberID(id2),
					"mazrean2",
					values.TrapMemberStatusActive,
				),
			},
			users: []*openapi.User{
				{
					Id:   id1.String(),
					Name: "mazrean",
				},
				{
					Id:   id2.String(),
					Name: "mazrean2",
				},
			},
		},
		{
			// 実際にはmiddlewareで弾かれるが、念の為確認
			description:  "sessionが存在しないのでauthSessionも存在せず500",
			sessionExist: false,
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			// 実際にはmiddlewareで弾かれるが、念の為確認
			description:      "authSessionが存在しないので500",
			sessionExist:     true,
			authSessionExist: false,
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
		{
			description:             "GetAllActiveUserがエラーなので500",
			sessionExist:            true,
			authSessionExist:        true,
			accessToken:             "accessToken",
			expiresAt:               time.Now(),
			executeGetAllActiveUser: true,
			GetAllActiveUserErr:     errors.New("error"),
			isErr:                   true,
			statusCode:              http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/users/me", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.New(req)
				if err != nil {
					t.Fatal(err)
				}

				if testCase.authSessionExist {
					sess.Values[accessTokenSessionKey] = testCase.accessToken
					sess.Values[expiresAtSessionKey] = testCase.expiresAt
				}

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)

				sess, err = session.Get(req)
				if err != nil {
					t.Fatal(err)
				}

				setCookieHeader(c)
			}

			if testCase.executeGetAllActiveUser {
				mockUserService.
					EXPECT().
					GetAllActiveUser(gomock.Any(), gomock.Any()).
					Return(testCase.userInfos, testCase.GetAllActiveUserErr)
			}

			users := userHandler.GetUsers(c)

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
			if err != nil {
				return
			}

			var resUser openapi.User
			assert.Equal(t, len(testCase.users), len(resUser))
			for i, user := range resUser {
				assert.Equal(t, *testCase.users[i], *user)
			}
		})
	}
}
