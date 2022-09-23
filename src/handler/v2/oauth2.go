package v2

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type OAuth2 struct {
	oauth2Unimplemented
	baseURL     *url.URL
	session     *Session
	oidcService service.OIDCV2
}

func NewOAuth2(conf config.Handler, session *Session, oidcService service.OIDCV2) (*OAuth2, error) {
	baseURL, err := conf.TraqBaseURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get traq base url: %v", err)
	}

	return &OAuth2{
		baseURL:     baseURL,
		session:     session,
		oidcService: oidcService,
	}, nil
}

// oauth2Unimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type oauth2Unimplemented interface {
	// traP Collectionの管理画面からのログアウト
	// (POST /oauth2/logout)
	PostLogout(ctx echo.Context) error
}

// traQのOAuth 2.0のコールバック
// (GET /oauth2/callback)
func (oauth2 *OAuth2) GetCallback(c echo.Context, params openapi.GetCallbackParams) error {
	code := values.NewOIDCAuthorizationCode(params.Code)

	session, err := oauth2.session.get(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	strCodeVerifier, err := oauth2.session.getCodeVerifier(session)
	if errors.Is(err, ErrNoValue) {
		return echo.NewHTTPError(http.StatusBadRequest, "no code verifier")
	}
	if err != nil {
		log.Printf("error: failed to get code verifier: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get code verifier")
	}
	codeVerifier := values.NewOIDCCodeVerifierFromString(strCodeVerifier)

	authState := domain.NewOIDCAuthState(
		values.OIDCCodeChallengeMethodSha256,
		codeVerifier,
	)

	authSession, err := oauth2.oidcService.Callback(c.Request().Context(), authState, code)
	if errors.Is(err, service.ErrInvalidAuthStateOrCode) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid auth state or code")
	}
	if err != nil {
		log.Printf("error: failed to callback: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to callback")
	}

	oauth2.session.setAuthSession(session, authSession)

	err = oauth2.session.save(c, session)
	if err != nil {
		log.Printf("error: failed to save session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session")
	}

	return c.NoContent(http.StatusOK)
}

// OAuth 2.0のCode Verifierなどのセッションへの設定とtraQへのリダイレクト
// (GET /oauth2/code)
func (oauth2 *OAuth2) GetCode(c echo.Context) error {
	client, authState, err := oauth2.oidcService.GenerateAuthState(c.Request().Context())
	if err != nil {
		log.Printf("error: failed to generate code: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate code")
	}

	codeChallenge, err := authState.GetCodeVerifier().GetCodeChallenge(authState.GetCodeChallengeMethod())
	if err != nil {
		log.Printf("error: failed to get code challenge: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get code challenge")
	}

	var strCodeChallengeMethod string
	switch authState.GetCodeChallengeMethod() {
	case values.OIDCCodeChallengeMethodSha256:
		strCodeChallengeMethod = "S256"
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get code challenge method")
	}

	session, err := oauth2.session.get(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	oauth2.session.setCodeVerifier(session, string(authState.GetCodeVerifier()))

	err = oauth2.session.save(c, session)
	if err != nil {
		log.Printf("error: failed to save session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session")
	}

	redirectURL := *oauth2.baseURL
	redirectURL.Path += "/oauth2/authorize"
	q := redirectURL.Query()
	q.Set("code_challenge", string(codeChallenge))
	q.Set("code_challenge_method", strCodeChallengeMethod)
	q.Set("client_id", string(client.GetClientID()))
	q.Set("response_type", "code")
	redirectURL.RawQuery = q.Encode()

	c.Response().Header().Set("Location", redirectURL.String())

	return echo.NewHTTPError(http.StatusSeeOther, fmt.Sprintf("redirect to %s", redirectURL.String()))
}
