package v2

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/session"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestGetMe(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCService := mock.NewMockOIDCV2(ctrl)
	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := session.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	userHandler := NewUser(session, mockOIDCService)

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
				false,
			),
			user: &openapi.User{
				Id:   id1,
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
			c, req, rec := setupTestRequest(t, http.MethodPost, "/users/me", nil)

			if testCase.sessionExist {
				var authSession *domain.OIDCSession
				if testCase.authSessionExist {
					authSession = domain.NewOIDCSession(values.NewOIDCAccessToken(testCase.accessToken), testCase.expiresAt)
				}
				setTestSession(t, c, req, rec, session, authSession)
			}

			if testCase.executeGetMe {
				mockOIDCService.
					EXPECT().
					GetMe(gomock.Any(), gomock.Any()).
					Return(testCase.userInfo, testCase.GetMeErr)
			}

			err = userHandler.GetMe(c)

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

			var resUser *openapi.User
			err = json.NewDecoder(rec.Body).Decode(&resUser)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Equal(t, testCase.user.Id, resUser.Id)
			assert.Equal(t, testCase.user.Name, resUser.Name)
		})
	}
}

func TestGetUsers(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCService := mock.NewMockOIDCV2(ctrl)
	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := session.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	userHandler := NewUser(session, mockOIDCService)

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
		bot                     bool
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
					false,
				),
			},
			users: []*openapi.User{
				{
					Id:   id1,
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
					false,
				),
				service.NewUserInfo(
					values.NewTrapMemberID(id2),
					"mazrean2",
					values.TrapMemberStatusActive,
					false,
				),
			},
			users: []*openapi.User{
				{
					Id:   id1,
					Name: "mazrean",
				},
				{
					Id:   id2,
					Name: "mazrean2",
				},
			},
		},
		{
			description:             "botがtrueでも正しく動く",
			sessionExist:            true,
			authSessionExist:        true,
			accessToken:             "accessToken",
			expiresAt:               time.Now(),
			executeGetAllActiveUser: true,
			bot:                     true,
			userInfos: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"w4ma",
					values.TrapMemberStatusActive,
					false,
				),
			},
			users: []*openapi.User{
				{
					Id:   id1,
					Name: "w4ma",
				},
			},
		},
		{
			description:             "botがfalseでも正しく動く",
			sessionExist:            true,
			authSessionExist:        true,
			accessToken:             "accessToken",
			expiresAt:               time.Now(),
			executeGetAllActiveUser: true,
			bot:                     false,
			userInfos: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"w4ma",
					values.TrapMemberStatusActive,
					false,
				),
			},
			users: []*openapi.User{
				{
					Id:   id1,
					Name: "w4ma",
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
			c, req, rec := setupTestRequest(t, http.MethodPost, "/users/me", nil)
			params := openapi.GetUsersParams{
				Bot: &testCase.bot,
			}

			if testCase.sessionExist {
				var authSession *domain.OIDCSession
				if testCase.authSessionExist {
					authSession = domain.NewOIDCSession(values.NewOIDCAccessToken(testCase.accessToken), testCase.expiresAt)
				}
				setTestSession(t, c, req, rec, session, authSession)
			}

			if testCase.executeGetAllActiveUser {
				mockOIDCService.
					EXPECT().
					GetActiveUsers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(testCase.userInfos, testCase.GetAllActiveUserErr)
			}

			err := userHandler.GetUsers(c, params)

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

			var resUsers []*openapi.User
			err = json.NewDecoder(rec.Body).Decode(&resUsers)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Equal(t, testCase.users, resUsers) // kokohuann
			for i, user := range resUsers {
				assert.Equal(t, user.Id, resUsers[i].Id)
				assert.Equal(t, user.Name, resUsers[i].Name)
			}
		})
	}
}
