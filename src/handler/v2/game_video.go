package v2

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/mazrean/formstream"
	echoform "github.com/mazrean/formstream/echo"
	"github.com/traPtitech/trap-collection-server/src/domain"
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
	parser, err := echoform.NewParser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	var (
		noContent = true
		video     *domain.GameVideo
		mime      openapi.GameVideoMime
	)
	err = parser.Register("content", func(file io.Reader, _ formstream.Header) error {
		noContent = false

		var err error
		video, err = gameVideo.gameVideoService.SaveGameVideo(c.Request().Context(), file, values.NewGameIDFromUUID(gameID))
		if errors.Is(err, service.ErrInvalidGameID) {
			return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
		}
		if errors.Is(err, service.ErrInvalidFormat) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid video file type")
		}
		if err != nil {
			log.Printf("error: failed to save game video: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to save game video")
		}

		if video.GetType() == values.GameVideoTypeMp4 {
			mime = openapi.Videomp4
		} else {
			log.Printf("error: unknown game video type: %v\n", video.GetType())
			return echo.NewHTTPError(http.StatusInternalServerError, "unknown game video type")
		}

		return nil
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to register content")
	}

	if err := parser.Parse(); err != nil {
		return err
	}

	if noContent {
		return echo.NewHTTPError(http.StatusBadRequest, "no content")
	}

	return c.JSON(http.StatusCreated, openapi.GameVideo{
		Id:        openapi.GameVideoID(video.GetID()),
		Mime:      mime,
		CreatedAt: video.GetCreatedAt(),
	})
}

// ゲーム動画のバイナリの取得
// (GET /games/{gameID}/videos/{gameVideoID})
func (gameVideo *GameVideo) GetGameVideo(c echo.Context, gameID openapi.GameIDInPath, gameVideoID openapi.GameVideoIDInPath) error {
	tmpURL, err := gameVideo.gameVideoService.GetGameVideo(c.Request().Context(), values.NewGameIDFromUUID(gameID), values.NewGameVideoIDFromUUID(gameVideoID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if errors.Is(err, service.ErrInvalidGameVideoID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameVideoID")
	}
	if err != nil {
		log.Printf("error: failed to get game video: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game video")
	}

	return c.Redirect(http.StatusSeeOther, (*url.URL)(tmpURL).String())
}

// ゲーム動画のメタ情報の取得
// (GET /games/{gameID}/videos/{gameVideoID}/meta)
func (gameVideo *GameVideo) GetGameVideoMeta(ctx echo.Context, gameID openapi.GameIDInPath, gameVideoID openapi.GameVideoIDInPath) error {
	video, err := gameVideo.gameVideoService.GetGameVideoMeta(ctx.Request().Context(), values.NewGameIDFromUUID(gameID), values.NewGameVideoIDFromUUID(gameVideoID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if errors.Is(err, service.ErrInvalidGameVideoID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameVideoID")
	}
	if err != nil {
		log.Printf("error: failed to get game video meta: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game video meta")
	}

	var mime openapi.GameVideoMime
	if video.GetType() == values.GameVideoTypeMp4 {
		mime = openapi.Videomp4
	} else {
		log.Printf("error: unknown game video type: %v\n", video.GetType())
		return echo.NewHTTPError(http.StatusInternalServerError, "unknown game video type")
	}

	return ctx.JSON(http.StatusOK, openapi.GameVideo{
		Id:        openapi.GameVideoID(video.GetID()),
		Mime:      mime,
		CreatedAt: video.GetCreatedAt(),
	})
}
