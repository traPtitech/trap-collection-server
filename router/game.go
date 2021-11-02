package router

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
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
	v1 "github.com/traPtitech/trap-collection-server/src/handler/v1"
	"github.com/traPtitech/trap-collection-server/storage"
	"golang.org/x/sync/errgroup"
)

// Game gameの構造体
type Game struct {
	db      model.DBMeta
	storage storage.Storage
	oauth   base.OAuth
	*v1.GameRole
}

func newGame(db model.DBMeta, oauth base.OAuth, storage storage.Storage, gameRole *v1.GameRole) *Game {
	game := new(Game)

	game.db = db
	game.storage = storage
	game.oauth = oauth
	game.GameRole = gameRole

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
		return nil, errors.New("unexpected error occurred while getting access token")
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
	switch operatingSystem {
	case "win32":
		operatingSystem = "windows"
	case "darwin":
		operatingSystem = "mac"
	}

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

func (g *Game) getGameFileName(gameID string, fileType string) (string, error) {
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

// PostFile POST /games/:gameID/asset/urlの処理部分
func (g *Game) PostFile(gameID string, file multipartFile, fileType string) (*openapi.GameFile, error) {
	if !g.db.IsValidAssetType(fileType) {
		return nil, errors.New("invalid file type")
	}

	fileName, err := g.getGameFileName(gameID, fileType)
	if err != nil {
		return nil, fmt.Errorf("failed to get file name: %w", err)
	}

	eg := errgroup.Group{}
	eg.Go(func() error {
		return g.storage.Save(fileName, file)
	})

	hash := md5.New()
	var gameFile *openapi.GameFile
	eg.Go(func() error {
		var err error
		byteMd5 := hash.Sum(nil)
		strMd5 := hex.EncodeToString(byteMd5)

		gameFile, err = g.db.InsertGameFile(gameID, model.AssetType(fileType), strMd5)
		if err != nil {
			return fmt.Errorf("failed to insert file: %w", err)
		}

		return nil
	})

	fileBuf := bytes.NewBuffer(nil)
	mw := io.MultiWriter(hash, fileBuf)
	_, err = io.Copy(mw, file)
	if err != nil {
		return nil, fmt.Errorf("failed to make MultiWriter: %w", err)
	}

	err = eg.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return gameFile, nil
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
