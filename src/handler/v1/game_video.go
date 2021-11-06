package v1

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameVideo struct {
	gameVideoService service.GameVideo
}

func NewGameVideo(gameVideoService service.GameVideo) *GameVideo {
	return &GameVideo{
		gameVideoService: gameVideoService,
	}
}

func (gv *GameVideo) PostVideo(strGameID string, video multipart.File) error {
	ctx := context.Background()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	err = gv.gameVideoService.SaveGameVideo(ctx, video, gameID)
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if errors.Is(err, service.ErrInvalidFormat) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid video format")
	}
	if err != nil {
		log.Printf("error: failed to save video: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save video")
	}

	return nil
}

func (gv *GameVideo) GetVideo(strGameID string) (io.Reader, error) {
	ctx := context.Background()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	/*
		メモリに保持してしまうので、
		大きい画像を返すとメモリが溶けてしまう。
		Pipeを使いたいが、openapiでの生成コードからio.Writerが渡されておらず、
		エラーハンドリングが怪しくなるので一旦これで妥協する。
	*/
	buf := bytes.NewBuffer(nil)

	err = gv.gameVideoService.GetGameVideo(ctx, buf, gameID)
	if errors.Is(err, service.ErrNoGameVideo) {
		return nil, echo.NewHTTPError(http.StatusNotFound, "no video")
	}
	if errors.Is(err, service.ErrInvalidGameID) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if err != nil {
		log.Printf("error: failed to get video: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get video")
	}

	return buf, nil
}
