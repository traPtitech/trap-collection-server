package v2

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameVideo struct {
	gameVideoService service.GameVideoV2
}

func NewGameVideo(gameVideoService service.GameVideoV2) *GameVideo {
	return &GameVideo{
		gameVideoService: gameVideoService,
	}
}

// gameVideoUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameVideoUnimplemented interface {
	// ゲーム動画の作成
	// (GET /games/{gameID}/videos)
	GetGameVideos(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲーム動画一覧の取得
	// (POST /games/{gameID}/videos)
	PostGameVideo(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲーム動画のバイナリの取得
	// (GET /games/{gameID}/videos/{gameVideoID})
	GetGameVideo(ctx echo.Context, gameID openapi.GameIDInPath, gameVideoID openapi.GameVideoIDInPath) error
	// ゲーム動画のメタ情報の取得
	// (GET /games/{gameID}/videos/{gameVideoID}/meta)
	GetGameVideoMeta(ctx echo.Context, gameID openapi.GameIDInPath, gameVideoID openapi.GameVideoIDInPath) error
}

// ゲーム動画一覧の取得
// (GET /games/{gameID}/videos)
func (gameVideo *GameVideo) GetGameVideos(c echo.Context, gameID openapi.GameIDInPath) error {
	videos, err := gameVideo.gameVideoService.GetGameVideos(c.Request().Context(), values.NewGameIDFromUUID(gameID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if err != nil {
		log.Printf("error: failed to get game videos: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game videos")
	}

	resVideos := make([]openapi.GameVideo, 0, len(videos))
	for _, video := range videos {
		var mime openapi.GameVideoMime
		if video.GetType() == values.GameVideoTypeMp4 {
			mime = openapi.Videomp4
		} else {
			log.Printf("error: unknown game video type: %v\n", video.GetType())
			return echo.NewHTTPError(http.StatusInternalServerError, "unknown game video type")
		}

		resVideos = append(resVideos, openapi.GameVideo{
			Id:        openapi.GameVideoID(video.GetID()),
			Mime:      mime,
			CreatedAt: video.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusOK, resVideos)
}

// ゲーム動画の作成
// (POST /games/{gameID}/videos)
func (gameVideo *GameVideo) PostGameVideo(c echo.Context, gameID openapi.GameIDInPath) error {
	header, err := c.FormFile("content")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid file")
	}
	file, err := header.Open()
	if err != nil {
		log.Printf("error: failed to open file: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open file")
	}
	defer file.Close()

	video, err := gameVideo.gameVideoService.SaveGameVideo(c.Request().Context(), file, values.NewGameIDFromUUID(gameID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if err != nil {
		log.Printf("error: failed to save game video: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save game video")
	}

	var mime openapi.GameVideoMime
	if video.GetType() == values.GameVideoTypeMp4 {
		mime = openapi.Videomp4
	} else {
		log.Printf("error: unknown game video type: %v\n", video.GetType())
		return echo.NewHTTPError(http.StatusInternalServerError, "unknown game video type")
	}

	return c.JSON(http.StatusCreated, openapi.GameVideo{
		Id:        openapi.GameVideoID(video.GetID()),
		Mime:      mime,
		CreatedAt: video.GetCreatedAt(),
	})
}
