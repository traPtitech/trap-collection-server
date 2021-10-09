package auth

import "errors"

var (
	ErrIdpBroken          = errors.New("idp is broken")
	ErrInvalidClient      = errors.New("invalid client")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
