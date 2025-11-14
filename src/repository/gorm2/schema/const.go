package schema

const (
	ProductKeyStatusActive   = "active"   // 有効
	ProductKeyStatusInactive = "inactive" /// 無効
)

const (
	SeatStatusNone  = "none"
	SeatStatusEmpty = "empty"
	SeatStatusInUse = "in_use"
)

const (
	GameVisibilityTypePublic  = "public"
	GameVisibilityTypeLimited = "limited"
	GameVisibilityTypePrivate = "private"
)

const (
	GameFileTypeJar     = "jar"
	GameFileTypeWindows = "windows"
	GameFileTypeMac     = "mac"
)

const (
	GameImageTypeJpeg = "jpeg"
	GameImageTypePng  = "png"
	GameImageTypeGif  = "gif"
)

const (
	GameVideoTypeMp4 = "mp4"
	GameVideoTypeMkv = "mkv"
	GameVideoTypeM4v = "m4v"
)

const (
	GameManagementRoleTypeAdministrator = "administrator"
	GameManagementRoleTypeCollaborator  = "collaborator"
)
