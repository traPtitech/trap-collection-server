package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameFile struct {
	gameFileService service.GameFileV2
}

func NewGameFile(gameFileService service.GameFileV2) *GameFile {
	return &GameFile{
		gameFileService: gameFileService,
	}
}

// gameFileUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameFileUnimplemented interface {
	// ゲームファイルの作成
	// (GET /games/{gameID}/files)
	GetGameFiles(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲームファイル一覧の取得
	// (POST /games/{gameID}/files)
	PostGameFile(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲームファイルのバイナリの取得
	// (GET /games/{gameID}/files/{gameFileID})
	GetGameFile(ctx echo.Context, gameID openapi.GameIDInPath, gameFileID openapi.GameFileIDInPath) error
	// ゲームファイルのメタ情報の取得
	// (GET /games/{gameID}/files/{gameFileID}/meta)
	GetGameFileMeta(ctx echo.Context, gameID openapi.GameIDInPath, gameFileID openapi.GameFileIDInPath) error
}
