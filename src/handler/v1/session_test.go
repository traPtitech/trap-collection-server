package v1

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
)

// テスト用のレスポンスのSet-CookieヘッダーをCookieヘッダーに移す関数
func setCookieHeader(c echo.Context) {
	cookie := c.Response().Header().Get("Set-Cookie")
	c.Response().Header().Del("Set-Cookie")
	c.Request().Header.Set("Cookie", cookie)
}

func TestGetSession(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mock.NewMockHandler(ctrl)
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

	type test struct {
		description  string
		sessionExist bool
		isErr        bool
		err          error
	}

	testCases := []test{
		{
			description:  "特に問題ないのでエラーなし",
			sessionExist: true,
		},
		{
			description:  "セッションが存在しなくてもエラーなし",
			sessionExist: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.New(req)
				if err != nil {
					t.Fatal(err)
				}

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)
			}

			_, err := session.getSession(c)
			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
