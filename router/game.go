package router

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/h2non/filetype"
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
func (g *Game) PostGame(game *openapi.NewGame, c echo.Context) (*openapi.GameInfo, error) {
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

	gameInfo, err := g.db.PostGame(user.Id, game.Name, game.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to add game: %w", err)
	}

	return gameInfo, nil
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
func (g *Game) PutGame(gameID string, gameMeta *openapi.NewGame) (*openapi.GameInfo, error) {
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
	if err != nil {
		return nil, fmt.Errorf("failed to get file name: %w", err)
	}

	file, err := g.storage.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed In Opening Game File: %w", err)
	}

	return file, nil
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

// GetImage GET /games/:gameID/imageの処理部分
func (g *Game) GetImage(gameID string) (io.Reader, error) {
	imageFile, err := g.getIntroduction(gameID, "image")
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting Introduction File: %w", err)
	}

	return imageFile, nil
}

var imageExts []string = []string{"jpg", "png"}

// PostImage POST /game/:gameID/image
func (g *Game) PostImage(gameID string, image multipartFile) error {
	fileTypeBuf := bytes.NewBuffer(nil)
	fileBuf := bytes.NewBuffer(nil)
	mw := io.MultiWriter(fileTypeBuf, fileBuf)
	_, err := io.Copy(mw, image)
	if err != nil {
		return fmt.Errorf("failed to make MultiWriter: %w", err)
	}

	fileType, err := filetype.MatchReader(fileTypeBuf)
	if err != nil {
		return fmt.Errorf("failed to get filetype")
	}

	ext := fileType.Extension
	isValidExt := false
	for _, validExt := range imageExts {
		if ext == validExt {
			isValidExt = true
		}
	}
	if !isValidExt {
		return errors.New("invalid extension")
	}

	err = g.db.InsertIntroduction(gameID, "image", ext)
	if err != nil {
		return fmt.Errorf("failed to insert introduction: %w", err)
	}

	fileName := g.getImageFileName(gameID, ext)
	err = g.storage.Save(fileName, fileBuf)
	if err != nil {
		return fmt.Errorf("failed to save introduction: %w", err)
	}

	return nil
}

func (g *Game) getImageFileName(gameID string, ext string) string {
	return gameID + "_image." + ext
}

// GetVideo GET /games/:gameID/videoの処理部分
func (g *Game) GetVideo(gameID string) (io.Reader, error) {
	videoFile, err := g.getIntroduction(gameID, "video")
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting Introduction File: %w", err)
	}
	return videoFile, nil
}

var videoExts []string = []string{"mp4"}

// PostVideo POST /game/:gameID/video
func (g *Game) PostVideo(gameID string, video multipartFile) error {
	fileTypeBuf := bytes.NewBuffer(nil)
	fileBuf := bytes.NewBuffer(nil)
	mw := io.MultiWriter(fileTypeBuf, fileBuf)
	_, err := io.Copy(mw, video)
	if err != nil {
		return fmt.Errorf("failed to make MultiWriter: %w", err)
	}

	fileType, err := filetype.MatchReader(fileTypeBuf)
	if err != nil {
		return fmt.Errorf("failed to get filetype")
	}

	ext := fileType.Extension
	if ext == "m4v" {
		ext = "mp4"
	}
	isValidExt := false
	for _, validExt := range videoExts {
		if ext == validExt {
			isValidExt = true
			break
		}
	}
	if !isValidExt {
		return errors.New("invalid extension")
	}

	err = g.db.InsertIntroduction(gameID, "video", ext)
	if err != nil {
		return fmt.Errorf("failed to insert introduction: %w", err)
	}

	fileName := g.getVideoFileName(gameID, ext)
	err = g.storage.Save(fileName, fileBuf)
	if err != nil {
		return fmt.Errorf("failed to save introduction: %w", err)
	}

	return nil
}

func (g *Game) getVideoFileName(gameID string, ext string) string {
	return gameID + "_video." + ext
}

var roleMap = map[string]int8{
	"image": 0,
	"video": 1,
}

func (g *Game) getIntroduction(gameID string, role string) (io.Reader, error) {
	intRole, ok := roleMap[role]
	if !ok {
		return nil, errors.New("Invalid Role")
	}

	ext, err := g.db.GetExtension(gameID, intRole)
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting Extensions: %w", err)
	}

	var fileName string
	switch role {
	case "image":
		fileName = g.getImageFileName(gameID, ext)
	case "video":
		fileName = g.getVideoFileName(gameID, ext)
	}

	file, err := g.storage.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting File: %w", err)
	}

	return file, nil
}

// PostURL POST /games/:gameID/asset/urlの処理部分
func (g *Game) PostURL(gameID string, newGameURL *openapi.NewGameUrl) (*openapi.GameUrl, error) {
	gameURL, err := g.db.InsertGameURL(gameID, newGameURL.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to insert url: %w", err)
	}

	return gameURL, nil
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

// GetMaintainer GET /games/:gameID/maintainer
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

// PostGameVersion POST /games/:gameID/version
func (g *Game) PostGameVersion(gameID string, newGameVersion *openapi.NewGameVersion) (*openapi.GameVersion, error) {
	gameVersion, err := g.db.InsertGameVersion(gameID, newGameVersion.Name, newGameVersion.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to insert game version: %w", err)
	}

	return gameVersion, nil
}

// GetGameVersion /games/:gameID/version
func (g *Game) GetGameVersion(gameID string) ([]*openapi.GameVersion, error) {
	gameVersions, err := g.db.GetGameVersions(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game version: %w", err)
	}

	return gameVersions, nil
}
