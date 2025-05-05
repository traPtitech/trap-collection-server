package migrate

import "github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"

// アプリケーションのv1
type (
	GameTable                   = schema.GameTable2
	GameVersionTable            = schema.GameVersionTable
	GameURLTable                = schema.GameURLTable
	GameFileTable               = schema.GameFileTable
	GameFileTypeTable           = schema.GameFileTypeTable
	GameImageTable              = schema.GameImageTable
	GameImageTypeTable          = schema.GameImageTypeTable
	GameVideoTable              = schema.GameVideoTable
	GameVideoTypeTable          = schema.GameVideoTypeTable
	GameManagementRoleTable     = schema.GameManagementRoleTable
	GameManagementRoleTypeTable = schema.GameManagementRoleTypeTable
	LauncherVersionTable        = schema.LauncherVersionTable
	LauncherUserTable           = schema.LauncherUserTable
	LauncherSessionTable        = schema.LauncherSessionTable
)

const (
	GameFileTypeJar     = gameFileTypeJarV1
	GameFileTypeWindows = gameFileTypeWindowsV1
	GameFileTypeMac     = gameFileTypeMacV1
)

const (
	GameImageTypeJpeg = gameImageTypeJpegV1
	GameImageTypePng  = gameImageTypePngV1
	GameImageTypeGif  = gameImageTypeGifV1
)

const (
	GameVideoTypeMp4 = gameVideoTypeMp4V1
	GameVideoTypeMkv = gameVideoTypeMkvV14
	GameVideoTypeM4v = gameVideoTypeM4vV14
)

const (
	GameManagementRoleTypeAdministrator = gameManagementRoleTypeAdministratorV1
	GameManagementRoleTypeCollaborator  = gameManagementRoleTypeCollaboratorV1
)

// アプリケーションのv2
type (
	GameTable2              = schema.GameTable2 // 実際に使用されるテーブルはv1のGameTableと同一
	GameVersionTable2       = schema.GameVersionTable2
	GameFileTable2          = schema.GameFileTable2
	GameImageTable2         = schema.GameImageTable2
	GameVideoTable2         = schema.GameVideoTable2
	EditionTable2           = schema.EditionTable
	ProductKeyTable2        = schema.ProductKeyTable
	ProductKeyStatusTable2  = schema.ProductKeyStatusTable
	AccessTokenTable2       = schema.AccessTokenTable
	AdminTable              = schema.AdminTable
	SeatTable2              = schema.SeatTable
	SeatStatusTable2        = schema.SeatStatusTable
	GameGenreTable          = schema.GameGenreTable
	GameVisibilityTypeTable = schema.GameVisibilityTypeTable
)

const (
	ProductKeyStatusActive   = productKeyStatusActiveV6
	ProductKeyStatusInactive = productKeyStatusInactiveV6
)

const (
	SeatStatusNone  = seatStatusNoneV9
	SeatStatusEmpty = seatStatusEmptyV9
	SeatStatusInUse = seatStatusInUseV9
)

const (
	GameVisibilityTypePublic  = gameVisibilityTypePublicV11
	GameVisibilityTypeLimited = gameVisibilityTypeLimitedV11
	GameVisibilityTypePrivate = gameVisibilityTypePrivateV11
)
