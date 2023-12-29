package migrate

// アプリケーションのv1
type (
	GameTable                   = gameTable
	GameVersionTable            = gameVersionTable
	GameURLTable                = gameURLTable
	GameFileTable               = gameFileTable
	GameFileTypeTable           = gameFileTypeTable
	GameImageTable              = gameImageTable
	GameImageTypeTable          = gameImageTypeTable
	GameVideoTable              = gameVideoTable
	GameVideoTypeTable          = gameVideoTypeTable
	GameManagementRoleTable     = gameManagementRoleTable
	GameManagementRoleTypeTable = gameManagementRoleTypeTable
	LauncherVersionTable        = launcherVersionTable
	LauncherUserTable           = launcherUserTable
	LauncherSessionTable        = launcherSessionTable
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
)

const (
	GameManagementRoleTypeAdministrator = gameManagementRoleTypeAdministratorV1
	GameManagementRoleTypeCollaborator  = gameManagementRoleTypeCollaboratorV1
)

// アプリケーションのv2
type (
	GameTable2              = gameTable2V12 // 実際に使用されるテーブルはv1のGameTableと同一
	GameVersionTable2       = gameVersionTable2V5
	GameFileTable2          = gameFileTable2V5
	GameImageTable2         = gameImageTable2V2
	GameVideoTable2         = gameVideoTable2V2
	EditionTable2           = editionTableV6
	ProductKeyTable2        = productKeyTableV6
	ProductKeyStatusTable2  = productKeyStatusTableV6
	AccessTokenTable2       = accessTokenTableV2
	AdminTable              = adminTable
	SeatTable2              = seatTableV9
	SeatStatusTable2        = seatStatusTableV9
	GameGenreTable          = gameGenreTableV12
	GameVisibilityTypeTable = gameVisibilityTypeTableV11
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
