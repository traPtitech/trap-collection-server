package migrate

// アプリケーションのv1
type (
	GameTable                   = gameTableV1
	GameVersionTable            = gameVersionTableV1
	GameURLTable                = gameURLTableV1
	GameFileTable               = gameFileTableV1
	GameFileTypeTable           = gameFileTypeTableV1
	GameImageTable              = gameImageTableV1
	GameImageTypeTable          = gameImageTypeTableV1
	GameVideoTable              = gameVideoTableV1
	GameVideoTypeTable          = gameVideoTypeTableV1
	GameManagementRoleTable     = gameManagementRoleTableV1
	GameManagementRoleTypeTable = gameManagementRoleTypeTableV1
	LauncherVersionTable        = launcherVersionTableV1
	LauncherUserTable           = launcherUserTableV1
	LauncherSessionTable        = launcherSessionTableV1
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
	GameTable2        = GameTable2V2 // 実際に使用されるテーブルはv1のGameTableと同一
	GameVersionTable2 = GameVersionTable2V2
	GameFileTable2    = GameFileTable2V2
	GameImageTable2   = GameImageTable2V2
	GameVideoTable2   = GameVideoTable2V2
	EditionTable2     = EditionTableV2
	ProductKeyTable2  = ProductKeyTableV2
	AccessTokenTable2 = AccessTokenTableV2
)
