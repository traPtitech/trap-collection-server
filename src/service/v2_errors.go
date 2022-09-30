package service

import "errors"

var (
	ErrOverlapBetweenOwnersAndMaintainers = errors.New("overlap between owners and maintainers")
	ErrOverlapBetweenUserAndOwners        = errors.New("overlap between login user and owners")
)
