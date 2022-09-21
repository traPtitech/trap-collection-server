package v2

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type Checker struct{}

func NewChecker() *Checker {
	return &Checker{}
}

func (m *Checker) check(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	// 一時的に未実装のものはチェックなしで通す
	checkerMap := map[string]openapi3filter.AuthenticationFunc{
		openapi.TrapMemberAuthScopes:       m.noAuthChecker, // TODO: TrapMemberAuthChecker
		openapi.AdminAuthScopes:            m.noAuthChecker, // TODO: AdminAuthChecker
		openapi.GameOwnerAuthScopes:        m.noAuthChecker, // TODO: GameOwnerAuthChecker
		openapi.GameMaintainerAuthScopes:   m.noAuthChecker, // TODO: GameMaintainerAuthChecker
		openapi.EditionAuthScopes:          m.noAuthChecker, // TODO: EditionAuthChecker
		openapi.EditionGameAuthScopes:      m.noAuthChecker, // TODO: EditionGameAuthChecker
		openapi.EditionGameFileAuthScopes:  m.noAuthChecker, // TODO: EditionGameFileAuthChecker
		openapi.EditionGameImageAuthScopes: m.noAuthChecker, // TODO: EditionGameImageAuthChecker
		openapi.EditionGameVideoAuthScopes: m.noAuthChecker, // TODO: EditionGameVideoAuthChecker
		openapi.EditionIDAuthScopes:        m.noAuthChecker, // TODO: EditionIDAuthChecker
	}

	checker, ok := checkerMap[input.SecuritySchemeName]
	if !ok {
		return fmt.Errorf("unknown security scheme: %s", input.SecuritySchemeName)
	}

	return checker(ctx, input)
}

// noAuthChecker
// 認証なしで通すチェッカー
// TODO: noAuthChecker削除
func (m *Checker) noAuthChecker(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
	return nil
}
