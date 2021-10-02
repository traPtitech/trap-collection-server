package values

import (
	"net/url"

	"github.com/google/uuid"
)

type (
	LauncherVersionID               uuid.UUID
	LauncherVersionName             string
	LauncherVersionQuestionnaireURL *url.URL
	LauncherUserID                  uuid.UUID
	LauncherUserProductKey          string
	LauncherUserStatus              int
	LauncherSessionID               uuid.UUID
	LauncherSessionAccessToken      string
)

const (
	LauncherUserStatusActive LauncherUserStatus = iota
	LauncherUserStatusInactive
)
