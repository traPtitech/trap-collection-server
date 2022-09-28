package service

import "errors"

var (
	ErrOverlapBetweenOwnersAndMaintainers = errors.New("overlap between owners and maintainers")
)
