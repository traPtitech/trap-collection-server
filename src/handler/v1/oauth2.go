package v1

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type OAuth2 struct {
	session     *Session
	oidcService service.OIDC
}

func NewOAuth2(session *Session, oidcService service.OIDC) *OAuth2 {
	return &OAuth2{
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

func (o *OAuth2) GetGeneratedCode(c echo.Context) (*openapi.InlineResponse200, error) {
	client, authState, err := o.oidcService.Authorize(c.Request().Context())
	if err != nil {
		log.Printf("error: failed to generate code: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to generate code")
	}

	codeChallenge, err := authState.GetCodeVerifier().GetCodeChallenge(authState.GetCodeChallengeMethod())
	if err != nil {
		log.Printf("error: failed to get code challenge: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get code challenge")
	}

	var strCodeChallengeMethod string
	switch authState.GetCodeChallengeMethod() {
	case values.OIDCCodeChallengeMethodSha256:
		strCodeChallengeMethod = "S256"
	default:
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get code challenge method")
	}

	session, err := o.session.getSession(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	o.session.setCodeVerifier(session, string(authState.GetCodeVerifier()))

	err = o.session.save(c, session)
	if err != nil {
		log.Printf("error: failed to save session: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to save session")
	}

	return &openapi.InlineResponse200{
		CodeChallenge:       string(codeChallenge),
		CodeChallengeMethod: strCodeChallengeMethod,
		ClientId:            string(client.GetClientID()),
		ResponseType:        "code",
	}, nil
}

func (o *OAuth2) PostLogout(c echo.Context) error {
	session, err := o.session.getSession(c)
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
