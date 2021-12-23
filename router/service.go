package router

import (
	"github.com/traPtitech/trap-collection-server/openapi"
	v1 "github.com/traPtitech/trap-collection-server/src/handler/v1"
)

// NewAPI Apiのコンストラクタ
func NewAPI(newAPI *v1.API, env string, clientID string, clientSecret string) (*openapi.Api, error) {
	game := newGame(newAPI.Game, newAPI.GameRole, newAPI.GameImage, newAPI.GameVideo, newAPI.GameVersion, newAPI.GameFile, newAPI.GameURL)

	api := &openapi.Api{
		Middleware:      newAPI.Middleware,
		GameApi:         game,
		LauncherAuthApi: newAPI.LauncherAuth,
		Oauth2Api:       newAPI.OAuth2,
		UserApi:         newAPI.User,
		VersionApi:      newAPI.LauncherVersion,
	}

	return api, nil
}
