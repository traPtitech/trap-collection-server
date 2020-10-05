package router

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
	"github.com/traPtitech/trap-collection-server/storage"
)

// Game gameの構造体
type Game struct {
	db      model.DBMeta
	storage storage.Storage
	oauth   base.OAuth
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
func (g *Game) PostGame(game *openapi.NewGameMeta, c echo.Context) (*openapi.GameMeta, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return nil, fmt.Errorf("failed to get session:%w", err)
	}

	interfaceAccessToken, ok := sess.Values["accessToken"]
	if !ok {
		log.Println("error: unexpected getting access token error")
		return nil, errors.New("unexpected error occcured while getting access token")
	}

	accessToken, ok := interfaceAccessToken.(string)
	if !ok {
		log.Println("error: unexpected parsing access token error")
		return nil, errors.New("failed to parse access token")
	}

	user, err := g.oauth.GetMe(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to GetMe: %w", err)
	}

	gameMeta, err := g.db.PostGame(user.Id, game.Name, game.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to add game: %w", err)
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

// PutGame PUT /games/:gameID
func (g *Game) PutGame(gameID string, gameMeta *openapi.NewGameMeta) (*openapi.GameMeta, error) {
	game, err := g.db.UpdateGame(gameID, gameMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return game, nil
}

// GetGames GET /gamesの処理部分
func (g *Game) GetGames(all string, c echo.Context) ([]*openapi.Game, error) {
	var games []*openapi.Game
	var isAll bool
	var err error

	if len(all) != 0 {
		isAll, err = strconv.ParseBool(all)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bool: %w", err)
		}
	}

	if !isAll {
		sess, err := session.Get("sessions", c)
		if err != nil {
			return nil, fmt.Errorf("Failed In Getting Session:%w", err)
		}

		interfaceAccessToken, ok := sess.Values["accessToken"]
		if !ok {
			log.Println("unexpected getting access token error")
			return nil, errors.New("Failed In Getting Access Token")
		}

		accessToken, ok := interfaceAccessToken.(string)
		if !ok {
			log.Println("unexpected parsing access token error")
			return nil, errors.New("Failed In Parsing Access Token")
		}

		user, err := g.oauth.GetMe(accessToken)
		if err != nil {
			return nil, fmt.Errorf("GetMe Error: %w", err)
		}

		games, err = g.db.GetGames(user.Id)
		if err != nil {
			return nil, fmt.Errorf("Failed In Getting Games: %w", err)
		}
	} else {
		var err error
		games, err = g.db.GetGames()
		if err != nil {
			return nil, fmt.Errorf("Failed In Getting Games: %w", err)
		}
	}

	return games, nil
}

// DeleteGames DELETE /games/:gameIDの処理部分
func (g *Game) DeleteGames(gameID string) error {
	err := g.db.DeleteGame(gameID)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	return nil
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
	"jar":     "jar",
	"windows": "zip",
	"mac":     "zip",
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
	var roleMap = map[string]int8{
		"image": 0,
		"video": 1,
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

// PostMaintainer POST /games/:gameID/maintainerの処理部分
func (g *Game) PostMaintainer(gameID string, maintainers *openapi.Maintainers, c echo.Context) error {
	userIDs := maintainers.Maintainers

	sess, err := session.Get("sessions", c)
	if err != nil {
		return fmt.Errorf("failed to get session:%w", err)
	}

	interfaceAccessToken, ok := sess.Values["accessToken"]
	if !ok {
		log.Println("error: unexpected getting access token error")
		return errors.New("unexpected error occcured while getting access token")
	}

	accessToken, ok := interfaceAccessToken.(string)
	if !ok {
		log.Println("error: unexpected parsing access token error")
		return errors.New("failed to parse access token")
	}

	users, err := g.oauth.GetUsers(accessToken)
	if err != nil {
		return fmt.Errorf("failed to GetUsers: %w", err)
	}

	userMap := make(map[string]*openapi.User, len(users))
	for _, user := range users {
		userMap[user.Id] = user
	}

	for _, userID := range userIDs {
		_, ok := userMap[userID]
		if !ok {
			return fmt.Errorf("invalid userID(%s)", userID)
		}
	}

	err = g.db.InsertMaintainer(gameID, userIDs)
	if err != nil {
		return fmt.Errorf("failed to insert maintainers: %w", err)
	}

	return nil
}

func (g *Game) GetMaintainer(gameID string, c echo.Context) ([]*openapi.Maintainer, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return nil, fmt.Errorf("failed to get session:%w", err)
	}

	interfaceAccessToken, ok := sess.Values["accessToken"]
	if !ok {
		log.Println("error: unexpected getting access token error")
		return nil, errors.New("unexpected error occcured while getting access token")
	}

	accessToken, ok := interfaceAccessToken.(string)
	if !ok {
		log.Println("error: unexpected parsing access token error")
		return nil, errors.New("failed to parse access token")
	}

	users, err := g.oauth.GetUsers(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to GetUsers: %w", err)
	}

	userMap := make(map[string]*openapi.User, len(users))
	for _, user := range users {
		userMap[user.Id] = user
	}

	maintainers, err := g.db.GetMaintainers(gameID, userMap)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintainers: %w", err)
	}

	return maintainers, nil
}

func (g *Game) GetGameVersion(gameID string) ([]*openapi.GameVersion, error) {
	gameVersions, err := g.db.GetGameVersions(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game version: %w", err)
	}

	return gameVersions, nil
}
