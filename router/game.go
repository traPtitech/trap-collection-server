package router

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
	v1 "github.com/traPtitech/trap-collection-server/src/handler/v1"
)

// Game gameの構造体
type Game struct {
	db    model.DBMeta
	oauth base.OAuth
	*v1.GameRole
	*v1.GameImage
	*v1.GameVideo
	*v1.GameVersion
	*v1.GameFile
	*v1.GameURL
}

func newGame(db model.DBMeta, oauth base.OAuth, gameRole *v1.GameRole, gameImage *v1.GameImage, gameVideo *v1.GameVideo, gameVersion *v1.GameVersion, gameFile *v1.GameFile, gameURL *v1.GameURL) *Game {
	game := new(Game)

	game.db = db
	game.oauth = oauth
	game.GameRole = gameRole
	game.GameImage = gameImage
	game.GameVideo = gameVideo
	game.GameVersion = gameVersion
	game.GameFile = gameFile
	game.GameURL = gameURL

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
