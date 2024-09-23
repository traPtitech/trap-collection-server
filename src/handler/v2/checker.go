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
	context                  *Context
	session                  *Session
	oidcService              service.OIDCV2
	editionService           service.Edition
	editionAuthService       service.EditionAuth
	gameRoleService          service.GameRoleV2
	administratorAuthService service.AdminAuthV2
}

func NewChecker(
	context *Context,
	session *Session,
	oidcService service.OIDCV2,
	editionService service.Edition,
	editionAuthService service.EditionAuth,
	gameRoleService service.GameRoleV2,
	administratorAuthService service.AdminAuthV2,
) *Checker {
	return &Checker{
		context:                  context,
		session:                  session,
		oidcService:              oidcService,
		editionService:           editionService,
		editionAuthService:       editionAuthService,
		gameRoleService:          gameRoleService,
		administratorAuthService: administratorAuthService,
	}
}

func (checker *Checker) check(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	// 一時的に未実装のものはチェックなしで通す
	checkerMap := map[string]openapi3filter.AuthenticationFunc{
		"TrapMemberAuth":       checker.TrapMemberAuthChecker,
		"AdminAuth":            checker.AdminAuthChecker,
		"GameOwnerAuth":        checker.GameOwnerAuthChecker,
		"GameMaintainerAuth":   checker.GameMaintainerAuthChecker,
		"EditionAuth":          checker.EditionAuthChecker,
		"EditionGameFileAuth":  checker.EditionGameFileAuthChecker,
		"EditionGameImageAuth": checker.EditionGameImageAuthChecker,
		"EditionGameVideoAuth": checker.EditionGameVideoAuthChecker,
		"EditionIDAuth":        checker.EditionIDAuthChecker,

		"GameInfoVisibilityAuth":  checker.NotImplementedChecker,
		"GameFileVisibilityAuth":  checker.NotImplementedChecker,
		"GameImageVisibilityAuth": checker.NotImplementedChecker,
		"GameVideoVisibilityAuth": checker.NotImplementedChecker,
	}

	checkerFunc, ok := checkerMap[input.SecuritySchemeName]
	if !ok {
		log.Printf("error: unknown security scheme: %s\n", input.SecuritySchemeName)
		return fmt.Errorf("unknown security scheme: %s", input.SecuritySchemeName)
	}

	err := checkerFunc(ctx, input)
	if err != nil {
		// fmt.Errorfでwrapすると*echo.HTTPError型でなくなりstatus codeが403になってしまうので、
		// ここではwrapしてはいけない
		return err
	}

	return nil
}

func (checker *Checker) NotImplementedChecker(context.Context, *openapi3filter.AuthenticationInput) error {
	return nil
}

// TrapMemberAuthChecker
// traPのメンバーかどうかをチェックするチェッカー
func (checker *Checker) TrapMemberAuthChecker(ctx context.Context, _ *openapi3filter.AuthenticationInput) error {
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

// AdminAuthChecker
// traPCollectionのadminであるかを調べるチェッカー
func (checker *Checker) AdminAuthChecker(ctx context.Context, _ *openapi3filter.AuthenticationInput) error {
	c := oapiMiddleware.GetEchoContext(ctx)
	// GetEchoContextの内部実装をみるとnilがかえりうるので、
	// ここではありえないはずだが念の為チェックする
	if c == nil {
		log.Printf("error: failed to get echo context\n")
		return errors.New("echo context is not set")
	}

	session, err := checker.session.get(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	authSession, err := checker.session.getAuthSession(session)
	if errors.Is(err, ErrNoValue) {
		return echo.NewHTTPError(http.StatusUnauthorized, "no access token")
	}

	err = checker.administratorAuthService.AdminAuthorize(ctx, authSession)
	if errors.Is(err, service.ErrOIDCSessionExpired) {
		return echo.NewHTTPError(http.StatusUnauthorized, "session is expired")
	}
	if errors.Is(err, service.ErrForbidden) {
		return echo.NewHTTPError(http.StatusUnauthorized, "not admin")
	}
	if err != nil {
		log.Printf("error: failed to authorize admin: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to authorize admin")
	}

	return nil
}

// GameOwnerAuthChecker
// そのゲームのowner(administrator)であるかどうかを調べるチェッカー
func (checker *Checker) GameOwnerAuthChecker(ctx context.Context, _ *openapi3filter.AuthenticationInput) error {
	c := oapiMiddleware.GetEchoContext(ctx)
	// GetEchoContextの内部実装をみるとnilがかえりうるので、
	// ここではありえないはずだが念の為チェックする
	if c == nil {
		log.Printf("error: failed to get echo context\n")
		return errors.New("echo context is not set")
	}

	session, err := checker.session.get(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	authSession, err := checker.session.getAuthSession(session)
	if errors.Is(err, ErrNoValue) {
		return echo.NewHTTPError(http.StatusUnauthorized, "no access token")
	}
	if err != nil {
		log.Printf("error: failed to get auth session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	//ランチャーの管理者は通す
	err = checker.administratorAuthService.AdminAuthorize(c.Request().Context(), authSession)
	if errors.Is(err, service.ErrOIDCSessionExpired) {
		return echo.NewHTTPError(http.StatusUnauthorized, "session is expired")
	}
	if err != nil && !errors.Is(err, service.ErrForbidden) {
		log.Printf("error: failed to check launcher admin auth: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check launcher admin auth")
	}
	if err == nil {
		return nil
	}

	strGameID := c.Param("gameID")
	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid gameID")
	}
	gameID := values.NewGameIDFromUUID(uuidGameID)

	err = checker.gameRoleService.UpdateGameManagementRoleAuth(ctx, authSession, gameID)
	if errors.Is(err, service.ErrForbidden) {
		return echo.NewHTTPError(http.StatusUnauthorized, "forbidden: not owner")
	}
	if errors.Is(err, service.ErrNoGame) {
		return echo.NewHTTPError(http.StatusNotFound, "no game")
	}
	if err != nil {
		log.Printf("error: failed to authorize game owner: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to authorize game owner")
	}
	return nil
}

// GameMaintainerAuthChecker
// そのゲームのmaintainer(collaborator)であるかどうかを調べるチェッカー
func (checker *Checker) GameMaintainerAuthChecker(ctx context.Context, _ *openapi3filter.AuthenticationInput) error {
	c := oapiMiddleware.GetEchoContext(ctx)
	// GetEchoContextの内部実装をみるとnilがかえりうるので、
	// ここではありえないはずだが念の為チェックする
	if c == nil {
		log.Printf("error: failed to get echo context\n")
		return errors.New("echo context is not set")
	}

	session, err := checker.session.get(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	authSession, err := checker.session.getAuthSession(session)
	if errors.Is(err, ErrNoValue) {
		return echo.NewHTTPError(http.StatusUnauthorized, "no access token")
	}
	if err != nil {
		log.Printf("error: failed to get auth session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	//ランチャーの管理者は通す
	err = checker.administratorAuthService.AdminAuthorize(c.Request().Context(), authSession)
	if errors.Is(err, service.ErrOIDCSessionExpired) {
		return echo.NewHTTPError(http.StatusUnauthorized, "session is expired")
	}
	if err != nil && !errors.Is(err, service.ErrForbidden) {
		log.Printf("error: failed to check launcher admin auth: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check launcher admin auth")
	}
	if err == nil {
		return nil
	}

	strGameID := c.Param("gameID")
	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid gameID")
	}
	gameID := values.NewGameIDFromUUID(uuidGameID)

	err = checker.gameRoleService.UpdateGameAuth(ctx, authSession, gameID)
	if errors.Is(err, service.ErrForbidden) {
		return echo.NewHTTPError(http.StatusUnauthorized, "forbidden: neither owner nor maintainer")
	}
	if errors.Is(err, service.ErrNoGame) {
		return echo.NewHTTPError(http.StatusNotFound, "no game")
	}
	if err != nil {
		log.Printf("error: failed to authorize game owner or maintainer: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to authorize game owner")
	}
	return nil
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

func (checker *Checker) EditionGameFileAuthChecker(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
	c := oapiMiddleware.GetEchoContext(ctx)
	// GetEchoContextの内部実装をみるとnilがかえりうるので、
	// ここではありえないはずだが念の為チェックする
	if c == nil {
		log.Printf("error: failed to get echo context\n")
		return errors.New("echo context is not set")
	}

	accessToken, ok, message := checker.getAccessToken(ai)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, message)
	}

	strFileID := c.Param("gameFileID")
	uuidFileID, err := uuid.Parse(strFileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid fileID")
	}
	fileID := values.NewGameFileIDFromUUID(uuidFileID)

	productKey, edition, err := checker.editionAuthService.EditionFileAuth(ctx, accessToken, fileID)
	if errors.Is(err, service.ErrInvalidAccessToken) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid access token")
	}
	if errors.Is(err, service.ErrExpiredAccessToken) {
		return echo.NewHTTPError(http.StatusUnauthorized, "expired access token")
	}
	if errors.Is(err, service.ErrForbidden) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	if err != nil {
		log.Printf("error: failed to check edition game file auth: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check edition game file auth")
	}

	checker.context.SetProductKey(c, productKey)
	checker.context.SetEdition(c, edition)

	return nil
}

func (checker *Checker) EditionGameImageAuthChecker(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
	c := oapiMiddleware.GetEchoContext(ctx)
	// GetEchoContextの内部実装をみるとnilがかえりうるので、
	// ここではありえないはずだが念の為チェックする
	if c == nil {
		log.Printf("error: failed to get echo context\n")
		return errors.New("echo context is not set")
	}

	accessToken, ok, message := checker.getAccessToken(ai)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, message)
	}

	strImageID := c.Param("gameImageID")
	uuidImageID, err := uuid.Parse(strImageID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid imageID")
	}
	imageID := values.GameImageIDFromUUID(uuidImageID)

	productKey, edition, err := checker.editionAuthService.EditionImageAuth(ctx, accessToken, imageID)
	if errors.Is(err, service.ErrInvalidAccessToken) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid access token")
	}
	if errors.Is(err, service.ErrExpiredAccessToken) {
		return echo.NewHTTPError(http.StatusUnauthorized, "expired access token")
	}
	if errors.Is(err, service.ErrForbidden) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	if err != nil {
		log.Printf("error: failed to check edition game image auth: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check edition game image auth")
	}

	checker.context.SetProductKey(c, productKey)
	checker.context.SetEdition(c, edition)

	return nil
}

func (checker *Checker) EditionGameVideoAuthChecker(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
	c := oapiMiddleware.GetEchoContext(ctx)
	// GetEchoContextの内部実装をみるとnilがかえりうるので、
	// ここではありえないはずだが念の為チェックする
	if c == nil {
		log.Printf("error: failed to get echo context\n")
		return errors.New("echo context is not set")
	}

	accessToken, ok, message := checker.getAccessToken(ai)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, message)
	}

	strVideoID := c.Param("gameVideoID")
	uuidVideoID, err := uuid.Parse(strVideoID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid videoID")
	}
	videoID := values.NewGameVideoIDFromUUID(uuidVideoID)

	productKey, edition, err := checker.editionAuthService.EditionVideoAuth(ctx, accessToken, videoID)
	if errors.Is(err, service.ErrInvalidAccessToken) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid access token")
	}
	if errors.Is(err, service.ErrExpiredAccessToken) {
		return echo.NewHTTPError(http.StatusUnauthorized, "expired access token")
	}
	if errors.Is(err, service.ErrForbidden) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	if err != nil {
		log.Printf("error: failed to check edition game video auth: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check edition game video auth")
	}

	checker.context.SetProductKey(c, productKey)
	checker.context.SetEdition(c, edition)

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
	accessToken, ok, message := checker.getAccessToken(ai)
	if !ok {
		return nil, nil, false, message, nil
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

func (checker *Checker) getAccessToken(ai *openapi3filter.AuthenticationInput) (values.LauncherSessionAccessToken, bool, string) {
	authorizationHeader := ai.RequestValidationInput.Request.Header.Get(echo.HeaderAuthorization)

	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return "", false, "invalid authorization header"
	}

	strAccessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")
	accessToken := values.NewLauncherSessionAccessTokenFromString(strAccessToken)
	if err := accessToken.Validate(); err != nil {
		return "", false, "invalid access token"
	}

	return accessToken, true, ""
}
