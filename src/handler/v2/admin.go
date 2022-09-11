package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type Admin struct {
	adminUnimplemented
}

func NewAdmin() *Admin {
	return &Admin{}
}

// adminUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type adminUnimplemented interface {
	// traPの管理者一覧取得
	// (GET /admins)
	GetAdmins(ctx echo.Context) error
	// traP Collection全体の管理者追加
	// (POST /admins)
	PostAdmin(ctx echo.Context) error
	// traP Collection全体の管理者削除
	// (DELETE /admins/{userID})
	DeleteAdmin(ctx echo.Context, userID openapi.UserIDInPath) error
}
