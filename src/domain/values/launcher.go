package values

import (
	"errors"
	"fmt"
	"net/url"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/pkg/random"
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

func NewLauncherVersionID(id uuid.UUID) LauncherVersionID {
	return LauncherVersionID(id)
}

func NewLauncherVersionName(name string) LauncherVersionName {
	return LauncherVersionName(name)
}

var (
	ErrLauncherVersionNameEmpty   = errors.New("launcher version name is empty")
	ErrLauncherVersionNameTooLong = errors.New("version name is too long")
)

func (lvn LauncherVersionName) Validate() error {
	// バージョン名は空ではない
	if len(lvn) == 0 {
		return ErrLauncherVersionNameEmpty
	}

	// バージョン名は32文字以内
	if utf8.RuneCountInString(string(lvn)) > 32 {
		return ErrLauncherVersionNameTooLong
	}

	return nil
}

func NewLauncherVersionQuestionnaireURL(url *url.URL) LauncherVersionQuestionnaireURL {
	return LauncherVersionQuestionnaireURL(url)
}

func NewLauncherUserID(id uuid.UUID) LauncherUserID {
	return LauncherUserID(id)
}

func NewLauncherUserProductKey() (LauncherUserProductKey, error) {
	randStr, err := random.SecureAlphaNumeric(25)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	key := fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		randStr[:5],
		randStr[5:10],
		randStr[10:15],
		randStr[15:20],
		randStr[20:25],
	)

	return LauncherUserProductKey(key), nil
}

func NewLauncherUserProductKeyFromString(key string) LauncherUserProductKey {
	return LauncherUserProductKey(key)
}

var (
	ErrLauncherUserProductKeyInvalidLength = errors.New("invalid length of product key")
	ErrLauncherUserProductKeyInvalidRune   = errors.New("invalid rune of product key")
)

func (lupk LauncherUserProductKey) Validate() error {
	if len(lupk) != 29 {
		return ErrLauncherUserProductKeyInvalidLength
	}

	for i, v := range lupk {
		if i == 5 || i == 11 || i == 17 || i == 23 {
			if v != '-' {
				return ErrLauncherUserProductKeyInvalidRune
			}
			continue
		}

		if !('0' <= v && v <= '9') && !('a' <= v && v <= 'z') && !('A' <= v && v <= 'Z') {
			return ErrLauncherUserProductKeyInvalidRune
		}
	}

	return nil
}

func NewLauncherSessionID(id uuid.UUID) LauncherSessionID {
	return LauncherSessionID(id)
}

func NewLauncherSessionAccessToken() (LauncherSessionAccessToken, error) {
	randStr, err := random.SecureAlphaNumeric(64)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	return LauncherSessionAccessToken(randStr), nil
}

func NewLauncherSessionAccessTokenFromString(token string) LauncherSessionAccessToken {
	return LauncherSessionAccessToken(token)
}

var (
	ErrLauncherSessionAccessTokenInvalidLength = errors.New("invalid length of access token")
	ErrLauncherSessionAccessTokenInvalidRune   = errors.New("invalid rune of access token")
)

func (lst LauncherSessionAccessToken) Validate() error {
	if len(lst) != 64 {
		return ErrLauncherSessionAccessTokenInvalidLength
	}

	for _, v := range lst {
		if !('0' <= v && v <= '9') && !('a' <= v && v <= 'z') && !('A' <= v && v <= 'Z') {
			return ErrLauncherSessionAccessTokenInvalidRune
		}
	}

	return nil
}
