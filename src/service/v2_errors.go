package service

import "errors"

var (
	ErrOverlapBetweenOwnersAndMaintainers = errors.New("overlap between owners and maintainers")
	ErrOverlapInOwners                    = errors.New("overlap (in owners/between login user and owners)")
	ErrOverlapInMaintainers               = errors.New("overlap in maintainers")
)
