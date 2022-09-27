package v2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Checker struct {
	session     *Session
	oidcService service.OIDCV2
}

func NewChecker(session *Session, oidcService service.OIDCV2) *Checker {
	return &Checker{
		session:     session,
		oidcService: oidcService,
	}
}

func (checker *Checker) check(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	// 一時的に未実装のものはチェックなしで通す
	checkerMap := map[string]openapi3filter.AuthenticationFunc{
		"TrapMemberAuth":       checker.TrapMemberAuthChecker,
		"AdminAuth":            checker.noAuthChecker, // TODO: AdminAuthChecker
		"GameOwnerAuth":        checker.noAuthChecker, // TODO: GameOwnerAuthChecker
		"GameMaintainerAuth":   checker.noAuthChecker, // TODO: GameMaintainerAuthChecker
		"EditionAuth":          checker.noAuthChecker, // TODO: EditionAuthChecker
		"EditionGameAuth":      checker.noAuthChecker, // TODO: EditionGameAuthChecker
		"EditionGameFileAuth":  checker.noAuthChecker, // TODO: EditionGameFileAuthChecker
		"EditionGameImageAuth": checker.noAuthChecker, // TODO: EditionGameImageAuthChecker
		"EditionGameVideoAuth": checker.noAuthChecker, // TODO: EditionGameVideoAuthChecker
		"EditionIDAuth":        checker.noAuthChecker, // TODO: EditionIDAuthChecker
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
