package v2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Checker struct {
	context            *Context
	session            *Session
	oidcService        service.OIDCV2
	editionService     service.Edition
	editionAuthService service.EditionAuth
}

func NewChecker(
	context *Context,
	session *Session,
	oidcService service.OIDCV2,
	editionService service.Edition,
	editionAuthService service.EditionAuth,
) *Checker {
	return &Checker{
		context:            context,
		session:            session,
		oidcService:        oidcService,
		editionService:     editionService,
		editionAuthService: editionAuthService,
	}
}

func (checker *Checker) check(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	// 一時的に未実装のものはチェックなしで通す
	checkerMap := map[string]openapi3filter.AuthenticationFunc{
		"TrapMemberAuth":       checker.TrapMemberAuthChecker,
		"AdminAuth":            checker.noAuthChecker, // TODO: AdminAuthChecker
		"GameOwnerAuth":        checker.noAuthChecker, // TODO: GameOwnerAuthChecker
		"GameMaintainerAuth":   checker.noAuthChecker, // TODO: GameMaintainerAuthChecker
		"EditionAuth":          checker.EditionAuthChecker,
		"EditionGameAuth":      checker.noAuthChecker, // TODO: EditionGameAuthChecker
		"EditionGameFileAuth":  checker.noAuthChecker, // TODO: EditionGameFileAuthChecker
		"EditionGameImageAuth": checker.noAuthChecker, // TODO: EditionGameImageAuthChecker
		"EditionGameVideoAuth": checker.noAuthChecker, // TODO: EditionGameVideoAuthChecker
		"EditionIDAuth":        checker.EditionIDAuthChecker,
	}

	checkerFunc, ok := checkerMap[input.SecuritySchemeName]
	if !ok {
		log.Printf("error: unknown security scheme: %s\n", input.SecuritySchemeName)
		return fmt.Errorf("unknown security scheme: %s", input.SecuritySchemeName)
	}

	return checkerFunc(ctx, input)
}

// noAuthChecker
// 認証なしで通すチェッカー
// TODO: noAuthChecker削除
func (checker *Checker) noAuthChecker(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
	return nil
}

// TrapMemberAuthChecker
// traPのメンバーかどうかをチェックするチェッカー
func (checker *Checker) TrapMemberAuthChecker(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
	c := oapiMiddleware.GetEchoContext(ctx)
	// GetEchoContextの内部実装をみるとnilがかえりうるので、
	// ここではありえないはずだが念の為チェックする
	if c == nil {
		log.Printf("error: failed to get echo context\n")
		return errors.New("echo context is not set")
	}

	ok, message, err := checker.checkTrapMemberAuth(c)
	if err != nil {
		log.Printf("error: failed to check launcher auth: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, message)
	}

	return nil
}

func (checker *Checker) checkTrapMemberAuth(c echo.Context) (bool, string, error) {
	session, err := checker.session.get(c)
	if err != nil {
		return false, "", fmt.Errorf("failed to get session: %w", err)
	}

	authSession, err := checker.session.getAuthSession(session)
	if errors.Is(err, ErrNoValue) {
		return false, "no access token", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("failed to get auth session: %w", err)
	}

	err = checker.oidcService.Authenticate(c.Request().Context(), authSession)
	if errors.Is(err, service.ErrOIDCSessionExpired) {
		return false, "access token is expired", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("failed to check traP auth: %w", err)
	}

	return true, "", nil
}

func (checker *Checker) EditionAuthChecker(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
	c := oapiMiddleware.GetEchoContext(ctx)
	// GetEchoContextの内部実装をみるとnilがかえりうるので、
	// ここではありえないはずだが念の為チェックする
	if c == nil {
		log.Printf("error: failed to get echo context\n")
		return errors.New("echo context is not set")
	}

	_, _, ok, message, err := checker.checkEditionAuth(c, ai)
	if err != nil {
		log.Printf("error: failed to check edition auth: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, message)
	}

	return nil
}

func (checker *Checker) EditionIDAuthChecker(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
	c := oapiMiddleware.GetEchoContext(ctx)
	// GetEchoContextの内部実装をみるとnilがかえりうるので、
	// ここではありえないはずだが念の為チェックする
	if c == nil {
		log.Printf("error: failed to get echo context\n")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	_, edition, ok, message, err := checker.checkEditionAuth(c, ai)
	if err != nil {
		log.Printf("error: failed to check edition auth: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, message)
	}

	strEditionID := c.Param("editionID")
	uuidEditionID, err := uuid.Parse(strEditionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid editionID")
	}
	editionID := values.NewLauncherVersionIDFromUUID(uuidEditionID)

	if editionID != edition.GetID() {
		return echo.NewHTTPError(http.StatusUnauthorized, "editionID is not matched")
	}

	return nil
}

func (checker *Checker) checkEditionAuth(c echo.Context, ai *openapi3filter.AuthenticationInput) (*domain.LauncherUser, *domain.LauncherVersion, bool, string, error) {
	authorizationHeader := ai.RequestValidationInput.Request.Header.Get(echo.HeaderAuthorization)

	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return nil, nil, false, "invalid authorization header", nil
	}

	strAccessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")
	accessToken := values.NewLauncherSessionAccessTokenFromString(strAccessToken)
	if err := accessToken.Validate(); err != nil {
		return nil, nil, false, "invalid access token", nil
	}

	productKey, edition, err := checker.editionAuthService.EditionAuth(c.Request().Context(), accessToken)
	if errors.Is(err, service.ErrInvalidAccessToken) {
		return nil, nil, false, "invalid access token", nil
	}
	if errors.Is(err, service.ErrExpiredAccessToken) {
		return nil, nil, false, "expired access token", nil
	}
	if err != nil {
		return nil, nil, false, "", fmt.Errorf("failed to check edition auth: %w", err)
	}

	checker.context.SetProductKey(c, productKey)
	checker.context.SetEdition(c, edition)

	return productKey, edition, true, "", nil
}
