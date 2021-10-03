package service

import "errors"

var (
	ErrInvalidLauncherVersion            = errors.New("invalid launcher version")
	ErrInvalidLauncherUser               = errors.New("invalid launcher user")
	ErrInvalidLauncherUserProductKey     = errors.New("invalid launcher user product key")
	ErrInvalidLauncherSessionAccessToken = errors.New("invalid launcher access token")
	ErrLauncherSessionAccessTokenExpired = errors.New("launcher access token expired")
)
