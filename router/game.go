package router

import (
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/storage"
)

// Game gameの構造体
type Game struct {
	storage.Storage
	openapi.GameApi
}

// NewGame Gemeのコンストラクタ
func NewGame(storage storage.Storage) *Game {
	game := &Game{
		Storage: storage,
	}
	return game
}

// GetGame GET /games/:gameID/infoの処理部分
func (*Game) GetGame(gameID string) (*openapi.Game, sessionMap, error) {
	game, err := model.GetGameInfo(gameID)
	if err != nil {
		return &openapi.Game{}, sessionMap{}, fmt.Errorf("Failed In Getting Game Info: %w", err)
	}
	return game, sessionMap{}, nil
}

// GetGameFile GET /games/asset/:gameID/fileの処理部分
func (g *Game) GetGameFile(gameID string, operatingSystem string) (ioReader, sessionMap, error) {
	fileName, err := g.getGameFileName(gameID, operatingSystem)
	file, err := g.Open(fileName)
	if err != nil {
		return nil, sessionMap{}, fmt.Errorf("Failed In Opening Game File: %w", err)
	}
	return file, sessionMap{}, nil
}

// GetImage GET /games/:gameID/imageの処理部分
func (g *Game) GetImage(gameID string) (ioReader, sessionMap, error) {
	imageFile, err := g.getIntroduction(gameID, "image")
	if err != nil {
		return nil, sessionMap{}, fmt.Errorf("Failed In Getting Introduction File: %w", err)
	}
	return imageFile, sessionMap{}, nil
}

// GetVideo GET /games/:gameID/videoの処理部分
func (g *Game) GetVideo(gameID string) (ioReader, sessionMap, error) {
	videoFile, err := g.getIntroduction(gameID, "video")
	if err != nil {
		return nil, sessionMap{}, fmt.Errorf("Failed In Getting Introduction File: %w", err)
	}
	return videoFile, sessionMap{}, nil
}

var typeExtMap map[string]string = map[string]string{
	"jar": "jar",
	"windows": "zip",
	"mac": "zip",
}

func (g *Game) getGameFileName(gameID string, operatingSystem string) (string, error) {
	fileType, err := model.GetGameType(gameID, operatingSystem)
	if err != nil {
		return "", fmt.Errorf("Failed In Getting Game Type: %w", err)
	}

	ext, ok := typeExtMap[fileType]
	if !ok {
		return "", errors.New("Invalid File Type")
	}

	return gameID + "_game." + ext, nil
}

func (g *Game) getIntroduction(gameID string, role string) (ioReader, error) {
	var roleMap = map[string]int8 {
		"image":0,
		"video":1,
	}

	intRole, ok := roleMap[role]
	if !ok {
		return nil, errors.New("Invalid Role")
	}

	ext, err := model.GetExtension(gameID, intRole)
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting Extensions: %w", err)
	}

	fileName := gameID + "_" + role + "." + ext
	file, err := g.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting File: %w", err)
	}

	return file, nil
}

// GetGameURL GET /games/:gameID/urlの処理部分
func (*Game) GetGameURL(gameID string) (string, sessionMap, error) {
	url, err := model.GetURL(gameID)
	if err != nil {
		return "", sessionMap{}, fmt.Errorf("Failed In Getting URL: %w", err)
	}

	return url, sessionMap{}, nil
}