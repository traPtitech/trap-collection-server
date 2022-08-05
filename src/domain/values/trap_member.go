package values

import (
	"errors"
	"unicode/utf8"

	"github.com/google/uuid"
)

type (
	TraPMemberID   uuid.UUID
	TraPMemberName string
	// 凍結されたかなどの状態
	TraPMemberStatus int
	// traP Collectionというアプリケーション上でのロール
	TraPMemberRole int
)

// traQのユーザーステータスより
// ref: https://apis.trap.jp/?urls.primaryName=traQ%20v3%20API のUserAccountState
const (
	TrapMemberStatusActive TraPMemberStatus = iota
	TrapMemberStatusDeactivated
	TrapMemberStatusSuspended

	// TrapMemberRoleUser
	// 通常ユーザー。
	TrapMemberRoleUser TraPMemberRole = iota
	// TrapMemberRoleAdmin
	// 管理者。非常時に対応できるように通常さわれないゲームの状態なども触れる。
	TrapMemberRoleAdmin
)

func NewTrapMemberID(id uuid.UUID) TraPMemberID {
	return TraPMemberID(id)
}

func NewTrapMemberName(name string) TraPMemberName {
	return TraPMemberName(name)
}

var (
	ErrTrapMemberNameEmpty       = errors.New("trap member name is empty")
	ErrTrapMemberNameTooLong     = errors.New("trap member name is too long")
	ErrTrapMemberNameInvalidRune = errors.New("trap member name contains invalid rune")
)

// Validate
// traQのtraQ IDのバリデーションに基づいて実装
// ref: https://github.com/traPtitech/traQ/blob/master/utils/validator/rules.go#L31-L35
func (tmn TraPMemberName) Validate() error {
	if len(tmn) == 0 {
		return ErrTrapMemberNameEmpty
	}

	if utf8.RuneCountInString(string(tmn)) > 32 {
		return ErrTrapMemberNameTooLong
	}

	for _, v := range tmn {
		if !('0' <= v && v <= '9') &&
			!('a' <= v && v <= 'z') &&
			!('A' <= v && v <= 'Z') &&
			v != '-' && v != '_' {
			return ErrTrapMemberNameInvalidRune
		}
	}

	return nil
}
