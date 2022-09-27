package v2

import "github.com/labstack/echo/v4"

type User struct {
	userUnimplemented
}

func NewUser() *User {
	return &User{}
}

// userUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type userUnimplemented interface {
	// traPのメンバー一覧取得
	// (GET /users)
	GetUsers(ctx echo.Context) error
	// ログイン中ユーザーの情報の取得
	// (GET /users/me)
	GetMe(ctx echo.Context) error
}
