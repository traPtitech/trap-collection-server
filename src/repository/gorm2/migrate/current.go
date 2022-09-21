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
	GameTable2        = gameTable2V2 // 実際に使用されるテーブルはv1のGameTableと同一
	GameVersionTable2 = gameVersionTable2V2
	GameFileTable2    = gameFileTable2V2
	GameImageTable2   = gameImageTable2V2
	GameVideoTable2   = gameVideoTable2V2
	EditionTable2     = editionTableV2
	ProductKeyTable2  = productKeyTableV2
	AccessTokenTable2 = accessTokenTableV2
)
