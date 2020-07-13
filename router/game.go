package router

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
	"github.com/traPtitech/trap-collection-server/storage"
)

// Game gameの構造体
type Game struct {
	db model.DBMeta
	storage storage.Storage
	oauth base.OAuth
	openapi.GameApi
}

func newGame(db model.DBMeta, oauth base.OAuth, storage storage.Storage) *Game {
	game := new(Game)

	game.db = db
	game.storage = storage
	game.oauth = oauth

	return game
}

//PostGame POST /gamesの処理部分
func (g *Game) PostGame(c echo.Context, game *openapi.NewGameMeta) (*openapi.GameMeta, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return nil, fmt.Errorf("failed to get session:%w", err)
	}

	interfaceAccessToken, ok := sess.Values["accessToken"]
	if !ok {
		log.Println("error: unexpected getting access token error")
		return nil, errors.New("Failed In Getting Access Token")
	}

	accessToken, ok := interfaceAccessToken.(string)
	if !ok {
		log.Println("error: unexpected parsing access token error")
		return nil, errors.New("Failed In Parsing Access Token")
	}

	user, err := g.oauth.GetMe(accessToken)
	if err != nil {
		return nil, fmt.Errorf("GetMe Error: %w", err)
	}

	gameMeta, err := g.db.PostGame(user.Id, game.Name, game.Description)
	if err != nil {
		return nil, fmt.Errorf("Failed In Adding Game: %w", err)
	}

	return gameMeta, nil
}

// GetGame GET /games/:gameID/infoの処理部分
func (g *Game) GetGame(gameID string) (*openapi.Game, error) {
	game, err := g.db.GetGameInfo(gameID)
	if err != nil {
		return &openapi.Game{}, fmt.Errorf("Failed In Getting Game Info: %w", err)
	}
	return game, nil
}

// GetGameFile GET /games/asset/:gameID/fileの処理部分
func (g *Game) GetGameFile(gameID string, operatingSystem string) (io.Reader, error) {
	fileName, err := g.getGameFileName(gameID, operatingSystem)
	file, err := g.storage.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed In Opening Game File: %w", err)
	}
	return file, nil
}

// GetImage GET /games/:gameID/imageの処理部分
func (g *Game) GetImage(gameID string) (io.Reader, error) {
	imageFile, err := g.getIntroduction(gameID, "image")
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting Introduction File: %w", err)
	}
	return imageFile, nil
}

// GetVideo GET /games/:gameID/videoの処理部分
func (g *Game) GetVideo(gameID string) (io.Reader, error) {
	videoFile, err := g.getIntroduction(gameID, "video")
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting Introduction File: %w", err)
	}
	return videoFile, nil
}

var typeExtMap map[string]string = map[string]string{
	"jar": "jar",
	"windows": "zip",
	"mac": "zip",
}

func (g *Game) getGameFileName(gameID string, operatingSystem string) (string, error) {
	fileType, err := g.db.GetGameType(gameID, operatingSystem)
	if err != nil {
		return "", fmt.Errorf("Failed In Getting Game Type: %w", err)
	}

	ext, ok := typeExtMap[fileType]
	if !ok {
		return "", errors.New("Invalid File Type")
	}

	return gameID + "_game." + ext, nil
}

func (g *Game) getIntroduction(gameID string, role string) (io.Reader, error) {
	var roleMap = map[string]int8 {
		"image":0,
		"video":1,
	}

	intRole, ok := roleMap[role]
	if !ok {
		return nil, errors.New("Invalid Role")
	}

	ext, err := g.db.GetExtension(gameID, intRole)
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting Extensions: %w", err)
	}

	fileName := gameID + "_" + role + "." + ext
	file, err := g.storage.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting File: %w", err)
	}

	return file, nil
}

// GetGameURL GET /games/:gameID/urlの処理部分
func (g *Game) GetGameURL(gameID string) (string, error) {
	url, err := g.db.GetURL(gameID)
	if err != nil {
		return "", fmt.Errorf("Failed In Getting URL: %w", err)
	}

	return url, nil
}
