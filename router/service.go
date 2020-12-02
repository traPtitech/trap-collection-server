package router

import (
	"fmt"
	"mime/multipart"
	"os"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
	"github.com/traPtitech/trap-collection-server/session"
	"github.com/traPtitech/trap-collection-server/storage"
)

type multipartFile = multipart.File

// Service serviceの構造体
type Service struct {
	*Middleware
	*Game
	*OAuth2
	*Seat
	*User
	*Version
}

// NewAPI Apiのコンストラクタ
func NewAPI(sess session.Session, env string, clientID string, clientSecret string) (*openapi.Api, error) {
	db := new(model.DB)

	var str storage.Storage
	if env == "development" || env == "mock" {
		localStr, err := storage.NewLocalStorage("./upload")
		if err != nil {
			return &openapi.Api{}, fmt.Errorf("Failed In LoacalStorage Constructor: %w", err)
		}
		str = localStr
	} else {
		swiftStr, err := storage.NewSwiftStorage(os.Getenv("container"))
		if err != nil {
			return &openapi.Api{}, fmt.Errorf("Failed In Swift Storage Constructor: %w", err)
		}
		str = swiftStr
	}

	strBaseURL := "https://q.trap.jp/api/v3"
	oauth, err := base.NewOAuth(strBaseURL)
	if err != nil {
		return &openapi.Api{}, fmt.Errorf("Failed In OAuth Constructor: %w", err)
	}

	launcherAuth := base.NewLauncherAuth()

	middleware := newMiddleware(db, oauth)
	game := newGame(db, oauth, str)
	oAuth2 := newOAuth2(sess, oauth, clientID, clientSecret)
	seat := newSeat(db, launcherAuth)
	user := newUser(oauth)
	version := newVersion(db, launcherAuth)

	api := &openapi.Api{
		Middleware: middleware,
		GameApi:    game,
		Oauth2Api:  oAuth2,
		SeatApi:    seat,
		UserApi:    user,
		VersionApi: version,
	}

	return api, nil
}
