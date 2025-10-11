package v2

import (
	"context"
	"net/http"
	"path"
	"slices"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/oapi-codegen/echo-middleware"
)

// isFileUploadRequest はファイルをアップロードするエンドポイントならtrueを返す。
// POST /api/v2/games/:gameID/{files,images,videos} へのリクエストが含まれる。
func isFileUploadRequest(c echo.Context) bool {
	if c.Request().Method != http.MethodPost {
		return false
	}

	gameID := c.Param("gameID")
	if gameID == "" {
		return false
	}

	targetPathBase := path.Join("/api/v2/games", gameID)
	targetPaths := []string{
		path.Join(targetPathBase, "files"),
		path.Join(targetPathBase, "images"),
		path.Join(targetPathBase, "videos"),
	}

	reqPath := path.Clean(c.Request().URL.Path)
	if slices.Contains(targetPaths, reqPath) {
		return true
	}

	return false
}

// fileUploadSkipper はファイルをアップロードするエンドポイントについてバリデーションをスキップする。
// OapiRequestValidator は 内部の ValidateSecurityRequirements でリクエストボディを全部読んでいる。
// そのため、画像・動画・ファイルのアップロード時にメモリ不足になる可能性がある。
// POST /api/v2/games/:gameID/{files,images,videos} へのリクエストは、バリデーションをスキップする。
func fileUploadSkipper(c echo.Context) bool {
	return isFileUploadRequest(c)
}

// fileUploadAuthMiddleware はファイルをアップロードするエンドポイントに対してのみ認証を行うミドルウェアを返す。
// POST /api/v2/games/:gameID/{files,images,videos} へのリクエストが含まれる。
//
// IMPORTANT: [(*Checker).GameMaintainerAuthChecker] の第2引数が使われていないことに依存した実装になっている。
func (checker *Checker) fileUploadAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !isFileUploadRequest(c) {
			return next(c)
		}

		// ref: [github.com/oapi-codegen/echo-middleware.GetEchoContext]
		// ここで echo.Context を context.Context に詰め込んでおかないと、GameMaintainerAuthChecker 内で取得できない。
		ctx := context.WithValue(c.Request().Context(), echomiddleware.EchoContextKey, c)

		err := checker.GameMaintainerAuthChecker(ctx, nil)
		if err != nil {
			return err
		}

		return next(c)
	}
}
