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
	// OAuth 2.0のCode Verifierなどのセッションへの設定とtraQへのリダイレクト
	// (GET /oauth2/code)
	GetCode(ctx echo.Context) error
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
