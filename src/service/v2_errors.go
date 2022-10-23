package service

import "errors"

var (
	ErrOverlapBetweenOwnersAndMaintainers = errors.New("overlap between owners and maintainers")
	ErrOverlapInOwners                    = errors.New("overlap (in owners/between login user and owners)")
	ErrOverlapInMaintainers               = errors.New("overlap in maintainers")
	ErrOffsetWithoutLimit                 = errors.New("there is offset but no limit")
	ErrCannotDeleteOwner                  = errors.New("cannot delete owner bacause there is only 1 owner")
	ErrCannotEditOwners                   = errors.New("cannot update role bacause there is only 1 owner")
	ErrNoAdminsUpdated                    = errors.New("no admins updated")
)
