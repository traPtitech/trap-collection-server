package v2

import (
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameGenre struct {
	gameGenreService       service.GameGenre
	gameGenreUnimplemented //実装し終わったら消す
}

func NewGameGenre(gameGenreService service.GameGenre) *GameGenre {
	return &GameGenre{
		gameGenreService: gameGenreService,
	}
}

// gameGenreUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameGenreUnimplemented interface {
	// ジャンル情報の変更
	// (PATCH /genres/{gameGenreID})
	PatchGameGenre(ctx echo.Context, gameGenreID openapi.GameGenreIDInPath) error
	// ゲームのジャンル編集
	// (PUT /games/{gameID}/genres)
	PutGameGenres(ctx echo.Context, gameID openapi.GameIDInPath) error
}

// ジャンルの削除
// (DELETE /genres/{gameGenreID})
func (gameGenre *GameGenre) DeleteGameGenre(c echo.Context, gameGenreID openapi.GameGenreIDInPath) error {
	err := gameGenre.gameGenreService.DeleteGameGenre(c.Request().Context(), values.GameGenreIDFromUUID(gameGenreID))
	if errors.Is(err, service.ErrNoGameGenre) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid game genre ID")
	}
	if err != nil {
		log.Printf("error: failed to delete game genre: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete game genre")
	}

	return c.NoContent(http.StatusOK)
}

// 全てのジャンルの取得
// (GET /genres)
func (gameGenre *GameGenre) GetGameGenres(ctx echo.Context) error {
	gameGenreInfos, err := gameGenre.gameGenreService.GetGameGenres(ctx.Request().Context())
	if err != nil {
		log.Printf("error: failed to get game genres: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game genres")
	}

	gameGenresResponse := make([]openapi.GameGenre, len(gameGenreInfos))
	for i := range gameGenreInfos {
		gameGenresResponse[i] = openapi.GameGenre{
			Id:        uuid.UUID(gameGenreInfos[i].GetID()),
			Genre:     string(gameGenreInfos[i].GetName()),
			Num:       gameGenreInfos[i].Num,
			CreatedAt: gameGenreInfos[i].GetCreatedAt(),
		}
	}

	return ctx.JSON(http.StatusOK, gameGenresResponse)
}
