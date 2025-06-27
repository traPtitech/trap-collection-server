package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
)

// setupTestRequestは、テスト用にEchoのContextを用意する。
// bodyOptがnilならば、リクエストボディは空になる。
func setupTestRequest(t *testing.T, method, path string, bodyOpt bodyOpt, opts ...echoContextOpt) (echo.Context, *http.Request, *httptest.ResponseRecorder) {
	t.Helper()

	var body io.Reader
	if bodyOpt != nil {
		var opt echoContextOpt
		body, opt = bodyOpt(t)
		opts = append(opts, opt)
	}

	req := httptest.NewRequestWithContext(t.Context(), method, path, body)
	for _, o := range opts {
		o(t, req)
	}

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c, req, rec
}

type echoContextOpt func(*testing.T, *http.Request)

type bodyOpt func(t *testing.T) (io.Reader, echoContextOpt)

func withJSONBody(t *testing.T, body any) bodyOpt {
	t.Helper()

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(body)
	require.NoError(t, err)

	return func(t *testing.T) (io.Reader, echoContextOpt) {
		t.Helper()

		return buf, func(t *testing.T, req *http.Request) {
			t.Helper()
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		}
	}
}

func withStringBody(t *testing.T, body string) bodyOpt {
	t.Helper()

	return func(t *testing.T) (io.Reader, echoContextOpt) {
		t.Helper()

		return bytes.NewBufferString(body), func(t *testing.T, req *http.Request) {
			t.Helper()
			req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
		}
	}
}

func withReaderBody(t *testing.T, body io.Reader, contentType string) bodyOpt {
	t.Helper()

	return func(t *testing.T) (io.Reader, echoContextOpt) {
		t.Helper()

		return body, func(t *testing.T, req *http.Request) {
			t.Helper()
			req.Header.Set(echo.HeaderContentType, contentType)
		}
	}
}

type testFormData struct {
	fieldName string
	fileName  string
	body      io.Reader
	value     string // フォームフィールド用の値
	isFile    bool   // ファイルかフォームフィールドかを判定
}

func withMultipartFormDataBody(t *testing.T, formDatas []testFormData) bodyOpt {
	t.Helper()

	return func(t *testing.T) (io.Reader, echoContextOpt) {
		t.Helper()

		reqBody := bytes.NewBuffer(nil)
		var boundary string

		func() {
			mw := multipart.NewWriter(reqBody)
			defer mw.Close()
			for _, d := range formDatas {
				if d.isFile && d.body != nil {
					w, err := mw.CreateFormFile(d.fieldName, d.fileName)
					require.NoError(t, err, "create form file")

					_, err = io.Copy(w, d.body)
					require.NoError(t, err, "copy to form file")
				} else if !d.isFile {
					w, err := mw.CreateFormField(d.fieldName)
					require.NoError(t, err, "create form field")

					_, err = w.Write([]byte(d.value))
					require.NoError(t, err, "write form field")
				}
			}
			boundary = mw.Boundary()
		}()

		return reqBody, func(t *testing.T, req *http.Request) {
			t.Helper()
			req.Header.Set(echo.HeaderContentType, fmt.Sprintf("%s; boundary=%s", echo.MIMEMultipartForm, boundary))
		}
	}
}

type sessionValue struct {
	key   string
	value any
}

// setTestSessionは、テスト用にCookieにセッションを設定する。
// authSessionがnilでなければ、OIDCセッションのアクセストークンと有効期限をセッションに保存する。
func setTestSession(t *testing.T, c echo.Context, req *http.Request, rec *httptest.ResponseRecorder,
	session *Session, authSession *domain.OIDCSession, sessionValues ...sessionValue) {
	t.Helper()

	sess, err := session.New(t, req)
	require.NoError(t, err, "create new session")

	if authSession != nil {
		sess.Values[accessTokenSessionKey] = string(authSession.GetAccessToken())
		sess.Values[expiresAtSessionKey] = authSession.GetExpiresAt()
	}

	for _, sv := range sessionValues {
		sess.Values[sv.key] = sv.value
	}

	err = sess.Save(req, rec)
	require.NoError(t, err, "save session")

	cookie := c.Response().Header().Get("Set-Cookie")
	c.Response().Header().Del("Set-Cookie")
	c.Request().Header.Set("Cookie", cookie)

	sess, err = session.Get(t, req)
	require.NoError(t, err, "get session")

	c.Set("session", sess)
}
