package v1

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type OAuth2 struct {
	baseURL     *url.URL
	session     *Session
	oidcService service.OIDC
}

func NewOAuth2(baseURL common.TraQBaseURL, session *Session, oidcService service.OIDC) *OAuth2 {
	return &OAuth2{
		baseURL:     baseURL,
		session:     session,
		oidcService: oidcService,
	}
}

func (o *OAuth2) Callback(strCode string, c echo.Context) error {
	code := values.NewOIDCAuthorizationCode(strCode)

	session, err := o.session.getSession(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	strCodeVerifier, err := o.session.getCodeVerifier(session)
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

	authSession, err := o.oidcService.Callback(c.Request().Context(), authState, code)
	if errors.Is(err, service.ErrInvalidAuthStateOrCode) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid auth state or code")
	}
	if err != nil {
		log.Printf("error: failed to callback: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to callback")
	}

	o.session.setAuthSession(session, authSession)

	err = o.session.save(c, session)
	if err != nil {
		log.Printf("error: failed to save session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session")
	}

	return nil
}

func (o *OAuth2) GetGeneratedCode(c echo.Context) error {
	client, authState, err := o.oidcService.Authorize(c.Request().Context())
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

	session, err := o.session.getSession(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	o.session.setCodeVerifier(session, string(authState.GetCodeVerifier()))

	err = o.session.save(c, session)
	if err != nil {
		log.Printf("error: failed to save session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session")
	}

	redirectURL := *o.baseURL
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

func (o *OAuth2) PostLogout(c echo.Context) error {
	session, err := getSession(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	authSession, err := o.session.getAuthSession(session)
	if errors.Is(err, ErrNoValue) {
		return echo.NewHTTPError(http.StatusBadRequest, "no auth session")
	}
	if err != nil {
		log.Printf("error: failed to get auth session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get auth session")
	}

	err = o.oidcService.Logout(c.Request().Context(), authSession)
	if err != nil {
		log.Printf("error: failed to logout: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to logout")
	}

	o.session.revoke(session)

	err = o.session.save(c, session)
	if err != nil {
		log.Printf("error: failed to save session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session")
	}

	return nil
}
