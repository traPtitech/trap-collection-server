package values

import (
	"net/url"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type (
	LauncherVersionID string
	/* Calender Versioning: https://calver.org/
	format: YYYY.0M.0D[-MODIFIER]*/
	LauncherVersionName string
	QuestionnaireURL string
	LauncherVersionCreatedAt time.Time
	LauncherVersionDeletedAt nullableTime
)

var (
	launcherVersionNameRegexp = regexp.MustCompile(`^\d{4}.[0-1]\d.[0-3]\d(|-[a-zA-Z0-9]+)$`)
	NullLauncherVersionDeletedAt LauncherVersionDeletedAt = LauncherVersionDeletedAt(nullTime)
)

func NewLuncherVersionID() LauncherVersionID {
	return LauncherVersionID(uuid.New().String())
}

func NewLauncherVersionIDFromString(id string) (LauncherVersionID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidFormat
	}

	return LauncherVersionID(id), nil
}

func NewLauncherVersionName(name string) (LauncherVersionName, error) {
	if !launcherVersionNameRegexp.MatchString(name) {
		return "", ErrInvalidFormat
	}

	return LauncherVersionName(name), nil
}

func NewQuestionnaireURL(u string) (QuestionnaireURL, error) {
	if urlObj, err := 	url.Parse(u); err != nil || !urlObj.IsAbs() {
		return "", ErrInvalidFormat
	}

	return QuestionnaireURL(u), nil
}

func NewLauncherVersionCreatedAt(createdAt time.Time) (LauncherVersionCreatedAt, error) {
	return LauncherVersionCreatedAt(createdAt), nil
}

func NewLauncherVersionDeletedAt(deletedAt time.Time) (LauncherVersionDeletedAt, error) {
	return LauncherVersionDeletedAt(nullableTime{
		time: deletedAt,
	}), nil
}
