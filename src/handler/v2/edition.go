package v2

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Edition struct {
	editionService service.Edition
	editionUnimplemented
}

func NewEdition(editionService service.Edition) *Edition {
	return &Edition{
		editionService: editionService,
	}
}

// editionUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type editionUnimplemented interface {
	// エディションの作成
	// (POST /editions)
	PostEdition(ctx echo.Context) error
	// エディションの削除
	// (DELETE /editions/{editionID})
	DeleteEdition(ctx echo.Context, editionID openapi.EditionIDInPath) error
	// エディション情報の取得
	// (GET /editions/{editionID})
	GetEdition(ctx echo.Context, editionID openapi.EditionIDInPath) error
	// エディション情報の変更
	// (PATCH /editions/{editionID})
	PatchEdition(ctx echo.Context, editionID openapi.EditionIDInPath) error
	// エディションに紐づくゲームの一覧の取得
	// (GET /editions/{editionID}/games)
	GetEditionGames(ctx echo.Context, editionID openapi.EditionIDInPath) error
	// エディションのゲームの変更
	// (PATCH /editions/{editionID}/games)
	PostEditionGame(ctx echo.Context, editionID openapi.EditionIDInPath) error
}

// エディション一覧の取得
// (GET /editions)
func (edition *Edition) GetEditions(c echo.Context) error {
	editions, err := edition.editionService.GetEditions(c.Request().Context())
	if err != nil {
		log.Printf("error: failed to get editions: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get editions")
	}

	res := make([]openapi.Edition, 0, len(editions))
	for _, edition := range editions {
		questionnaireURL, err := edition.GetQuestionnaireURL()
		if err != nil && !errors.Is(err, domain.ErrNoQuestionnaire) {
			log.Printf("error: failed to get questionnaire url: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get questionnaire url")
		}

		var strQuestionnaireURL *string
		if !errors.Is(err, domain.ErrNoQuestionnaire) {
			v := (*url.URL)(questionnaireURL).String()
			strQuestionnaireURL = &v
		}

		res = append(res, openapi.Edition{
			Id:            uuid.UUID(edition.GetID()),
			Name:          string(edition.GetName()),
			Questionnaire: strQuestionnaireURL,
			CreatedAt:     edition.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusOK, res)
}
