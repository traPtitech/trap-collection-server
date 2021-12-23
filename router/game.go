package router

import (
	v1 "github.com/traPtitech/trap-collection-server/src/handler/v1"
)

// Game gameの構造体
type Game struct {
	*v1.Game
	*v1.GameRole
	*v1.GameImage
	*v1.GameVideo
	*v1.GameVersion
	*v1.GameFile
	*v1.GameURL
}

func newGame(game *v1.Game, gameRole *v1.GameRole, gameImage *v1.GameImage, gameVideo *v1.GameVideo, gameVersion *v1.GameVersion, gameFile *v1.GameFile, gameURL *v1.GameURL) *Game {
	return &Game{
		Game:        game,
		GameRole:    gameRole,
		GameVersion: gameVersion,
		GameImage:   gameImage,
		GameVideo:   gameVideo,
		GameFile:    gameFile,
		GameURL:     gameURL,
	}
}
