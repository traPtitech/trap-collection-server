//go:build wireinject

package wire

import (
	"fmt"

	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/storage"
	"github.com/traPtitech/trap-collection-server/src/storage/local"
	"github.com/traPtitech/trap-collection-server/src/storage/swift"
)

var (
	storageSet = wire.NewSet(
		wire.FieldsOf(new(*Storage), "GameImage"),
		wire.FieldsOf(new(*Storage), "GameVideo"),
		wire.FieldsOf(new(*Storage), "GameFile"),

		storageSwitch,
	)
)

type Storage struct {
	GameImage storage.GameImage
	GameVideo storage.GameVideo
	GameFile  storage.GameFile
}

func newStorage(
	gameImage storage.GameImage,
	gameVideo storage.GameVideo,
	gameFile storage.GameFile,
) (*Storage, error) {
	return &Storage{
		GameImage: gameImage,
		GameVideo: gameVideo,
		GameFile:  gameFile,
	}, nil
}

func storageSwitch(
	conf config.Storage,
	swiftConf config.StorageSwift,
	localConf config.StorageLocal,
) (*Storage, error) {
	storageType, err := conf.Type()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage type: %w", err)
	}

	switch storageType {
	case config.StorageTypeSwift:
		return injectSwiftStorage(swiftConf)
	case config.StorageTypeLocal:
		return injectLocalStorage(localConf)
	}

	return nil, fmt.Errorf("unknown storage type: %d", storageType)
}

func injectSwiftStorage(conf config.StorageSwift) (*Storage, error) {
	wire.Build(
		wire.Bind(new(storage.GameImage), new(*swift.GameImage)),
		wire.Bind(new(storage.GameVideo), new(*swift.GameVideo)),
		wire.Bind(new(storage.GameFile), new(*swift.GameFile)),

		swift.NewClient,
		swift.NewGameImage,
		swift.NewGameVideo,
		swift.NewGameFile,

		newStorage,
	)

	return nil, nil
}

func injectLocalStorage(conf config.StorageLocal) (*Storage, error) {
	wire.Build(
		wire.Bind(new(storage.GameImage), new(*local.GameImage)),
		wire.Bind(new(storage.GameVideo), new(*local.GameVideo)),
		wire.Bind(new(storage.GameFile), new(*local.GameFile)),

		local.NewDirectoryManager,
		local.NewGameImage,
		local.NewGameVideo,
		local.NewGameFile,

		newStorage,
	)

	return nil, nil
}
