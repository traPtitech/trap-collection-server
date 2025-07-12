//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/src/handler"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
	// v1 "github.com/traPtitech/trap-collection-server/src/handler/v1"
	v2 "github.com/traPtitech/trap-collection-server/src/handler/v2"
)

var (
	handlerSet = wire.NewSet(
		common.NewSession,
		handler.NewAPI,
		// handlerV1Set,
		handlerV2Set,
	)
	// handlerV1Set = wire.NewSet(
	// 	v1.NewAPI,
	// 	v1.NewSession,
	// 	v1.NewGame,
	// 	v1.NewGameRole,
	// 	v1.NewGameImage,
	// 	v1.NewGameVideo,
	// 	v1.NewGameVersion,
	// 	v1.NewGameFile,
	// 	v1.NewGameURL,
	// 	v1.NewLauncherAuth,
	// 	v1.NewLauncherVersion,
	// 	v1.NewOAuth2,
	// 	v1.NewUser,
	// 	v1.NewMiddleware,
	// )
	handlerV2Set = wire.NewSet(
		v2.NewAPI,
		v2.NewChecker,
		v2.NewContext,
		v2.NewSession,
		v2.NewOAuth2,
		v2.NewUser,
		v2.NewAdmin,
		v2.NewGame,
		v2.NewGameRole,
		v2.NewGameGenre,
		v2.NewGameVersion,
		v2.NewGameFile,
		v2.NewGameImage,
		v2.NewGameVideo,
		v2.NewGamePlayLog,
		v2.NewEdition,
		v2.NewEditionAuth,
		v2.NewSeat,
	)
)
