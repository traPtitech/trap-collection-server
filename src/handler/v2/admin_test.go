package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestGetAdmins(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdminService := mock.NewMockAdminAuthV2(ctrl)

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

	adminHandler := NewAdmin(mockAdminService, session)

	type test struct {
		description      string
		sessionExist     bool
		authSession      *domain.OIDCSession
		executeGetAdmins bool
		GetAdminsErr     error
		adminInfos       []*service.UserInfo
		apiAdmins        []*openapi.User
		isErr            bool
		err              error
		statusCode       int
	}

	adminID1 := values.NewTrapMemberID(uuid.New())
	adminID2 := values.NewTrapMemberID(uuid.New())

	testCases := []test{
		{
			description:  "特に問題ないのでエラー無し",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetAdmins: true,
			adminInfos: []*service.UserInfo{
				service.NewUserInfo(adminID1, "ikura-hamu", values.TrapMemberStatusActive),
			},
			GetAdminsErr: nil,
			apiAdmins: []*openapi.User{
				{Id: uuid.UUID(adminID1), Name: "ikura-hamu"},
			},
			statusCode: http.StatusOK,
		},
		{
			description:  "sessionが無いので401",
			sessionExist: false,
			isErr:        true,
			statusCode:   http.StatusUnauthorized,
		},
		{
			description:  "auth session が無いので401",
			sessionExist: true,
			isErr:        true,
			statusCode:   http.StatusUnauthorized,
		},
		{
			description:  "GetAdminsがエラーなので500",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetAdmins: true,
			GetAdminsErr:     errors.New("error"),
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
		{
			description:  "adminが複数人いてもエラー無し",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			executeGetAdmins: true,
			GetAdminsErr:     nil,
			adminInfos: []*service.UserInfo{
				service.NewUserInfo(adminID1, "ikura-hamu", values.TrapMemberStatusActive),
				service.NewUserInfo(adminID2, "mazrean", values.TrapMemberStatusActive),
			},
			apiAdmins: []*openapi.User{
				{Id: uuid.UUID(adminID1), Name: "ikura-hamu"},
				{Id: uuid.UUID(adminID2), Name: "mazrean"},
			},
			statusCode: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/admins", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.New(req)
				if err != nil {
					t.Fatal(err)
				}

				if testCase.authSession != nil {
					sess.Values[accessTokenSessionKey] = string(testCase.authSession.GetAccessToken())
					sess.Values[expiresAtSessionKey] = testCase.authSession.GetExpiresAt()
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

				c.Set("session", sess)
			}

			if testCase.executeGetAdmins {
				mockAdminService.
					EXPECT().
					GetAdmins(gomock.Any(), gomock.Any()).
					Return(testCase.adminInfos, testCase.GetAdminsErr)
			}

			err := adminHandler.GetAdmins(c)

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

			var response []openapi.User
			err = json.NewDecoder(rec.Body).Decode(&response)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Equal(t, len(testCase.apiAdmins), len(response))

			for i, admin := range response {
				assert.Equal(t, testCase.apiAdmins[i].Id, admin.Id)
				assert.Equal(t, testCase.apiAdmins[i].Name, admin.Name)
			}
		})
	}
}

func TestPostAdmins(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdminService := mock.NewMockAdminAuthV2(ctrl)

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

	adminHandler := NewAdmin(mockAdminService, session)

	type test struct {
		description      string
		newAdminID       *openapi.PostAdminJSONBody
		sessionExist     bool
		authSession      *domain.OIDCSession
		executeAddAdmin  bool
		AddAdminErr      error
		adminInfos       []*service.UserInfo
		apiAdmins        []*openapi.User
		isBadRequestBody bool
		isErr            bool
		err              error
		statusCode       int
	}

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())

	testCases := []test{
		{
			description:  "特に問題ないのでエラー無し",
			newAdminID:   &openapi.PostAdminJSONBody{Id: uuid.UUID(userID1)},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeAddAdmin: true,
			adminInfos: []*service.UserInfo{
				service.NewUserInfo(userID1, "ikura-hamu", values.TrapMemberStatusActive),
			},
			apiAdmins: []*openapi.User{
				{Id: uuid.UUID(userID1), Name: "ikura-hamu"},
			},
		},
		{
			description:  "sessionが無いので401",
			sessionExist: false,
			isErr:        true,
			statusCode:   http.StatusUnauthorized,
		},
		{
			description:  "auth sessionが無いので401",
			sessionExist: true,
			isErr:        true,
			statusCode:   http.StatusUnauthorized,
		},
		{
			description:  "リクエストが不正なので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			isBadRequestBody: true,
			isErr:            true,
			statusCode:       http.StatusBadRequest,
		},
		{
			description:  "存在しないユーザーなので400",
			newAdminID:   &openapi.PostAdminJSONBody{Id: uuid.UUID(userID1)},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeAddAdmin: true,
			AddAdminErr:     service.ErrInvalidUserID,
			isErr:           true,
			statusCode:      http.StatusBadRequest,
		},
		{
			description:  "ユーザーが既にadminなので400",
			newAdminID:   &openapi.PostAdminJSONBody{Id: uuid.UUID(userID1)},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeAddAdmin: true,
			AddAdminErr:     service.ErrNoAdminsUpdated,
			isErr:           true,
			statusCode:      http.StatusBadRequest,
		},
		{
			description:  "AddAdminsがエラーなので500",
			newAdminID:   &openapi.PostAdminJSONBody{Id: uuid.UUID(userID1)},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeAddAdmin: true,
			AddAdminErr:     errors.New("test"),
			isErr:           true,
			statusCode:      http.StatusInternalServerError,
		},
		{
			description:  "他にadminがいてもエラー無し",
			newAdminID:   &openapi.PostAdminJSONBody{Id: uuid.UUID(userID1)},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeAddAdmin: true,
			adminInfos: []*service.UserInfo{
				service.NewUserInfo(userID2, "mazrean", values.TrapMemberStatusActive),
				service.NewUserInfo(userID1, "ikura-hamu", values.TrapMemberStatusActive),
			},
			apiAdmins: []*openapi.User{
				{Id: uuid.UUID(userID2), Name: "mazrean"},
				{Id: uuid.UUID(userID1), Name: "ikura-hamu"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			reqBody := new(bytes.Buffer)
			if !testCase.isBadRequestBody {
				err = json.NewEncoder(reqBody).Encode(testCase.newAdminID)
				if err != nil {
					log.Printf("failed to create request body")
					t.Fatal(err)
				}
			} else {
				reqBody = bytes.NewBufferString("bad requset body")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/admins", reqBody)
			req.Header.Set(echo.HeaderContentType, "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.New(req)
				if err != nil {
					t.Fatal(err)
				}

				if testCase.authSession != nil {
					sess.Values[accessTokenSessionKey] = string(testCase.authSession.GetAccessToken())
					sess.Values[expiresAtSessionKey] = testCase.authSession.GetExpiresAt()
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

				c.Set("session", sess)
			}

			if testCase.executeAddAdmin {
				mockAdminService.
					EXPECT().
					AddAdmin(gomock.Any(), gomock.Any(), values.NewTrapMemberID(testCase.newAdminID.Id)).
					Return(testCase.adminInfos, testCase.AddAdminErr)
			}

			err := adminHandler.PostAdmin(c)

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

			var response []openapi.User
			err = json.NewDecoder(rec.Body).Decode(&response)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Len(t, response, len(testCase.apiAdmins))

			for i, admin := range testCase.apiAdmins {
				assert.Equal(t, admin.Id, response[i].Id)
			}
		})
	}
}
