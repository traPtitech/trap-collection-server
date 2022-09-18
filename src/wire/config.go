//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/src/config"
	v1 "github.com/traPtitech/trap-collection-server/src/config/v1"
)

var configSet = wire.NewSet(
	wire.Bind(new(config.App), new(*v1.App)),
	v1.NewApp,

	wire.Bind(new(config.AuthTraQ), new(*v1.AuthTraQ)),
	v1.NewAuthTraQ,

	wire.Bind(new(config.CacheRistretto), new(*v1.CacheRistretto)),
	v1.NewCacheRistretto,

	wire.Bind(new(config.HandlerV1), new(*v1.HandlerV1)),
	v1.NewHandlerV1,

	wire.Bind(new(config.RepositoryGorm2), new(*v1.RepositoryGorm2)),
	v1.NewRepositoryGorm2,

	wire.Bind(new(config.ServiceV1), new(*v1.ServiceV1)),
	v1.NewServiceV1,

	wire.Bind(new(config.Storage), new(*v1.Storage)),
	v1.NewStorage,

	wire.Bind(new(config.StorageSwift), new(*v1.StorageSwift)),
	v1.NewStorageSwift,

	wire.Bind(new(config.StorageLocal), new(*v1.StorageLocal)),
	v1.NewStorageLocal,

	wire.Bind(new(config.StorageS3), new(*v1.StorageS3)),
	v1.NewStorageS3,
)
