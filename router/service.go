package router

import (
	"fmt"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
	v1 "github.com/traPtitech/trap-collection-server/src/handler/v1"
)

// Service serviceの構造体
type Service struct {
	*Game
	*Seat
	*Version
}

// NewAPI Apiのコンストラクタ
func NewAPI(newAPI *v1.API, env string, clientID string, clientSecret string) (*openapi.Api, error) {
	db := new(model.DB)

	strBaseURL := "https://q.trap.jp/api/v3"
	oauth, err := base.NewOAuth(strBaseURL)
	if err != nil {
		return &openapi.Api{}, fmt.Errorf("Failed In OAuth Constructor: %w", err)
	}

	launcherAuth := base.NewLauncherAuth()

	game := newGame(db, oauth, newAPI.GameRole, newAPI.GameImage, newAPI.GameVideo, newAPI.GameVersion, newAPI.GameFile, newAPI.GameURL)
	seat := newSeat(db, launcherAuth)
	version := newVersion(db, launcherAuth)

	api := &openapi.Api{
		Middleware:      newAPI.Middleware,
		GameApi:         game,
		LauncherAuthApi: newAPI.LauncherAuth,
		Oauth2Api:       newAPI.OAuth2,
		SeatApi:         seat,
		UserApi:         newAPI.User,
		VersionApi:      version,
	}

	return api, nil
}
