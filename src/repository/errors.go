package repository

import "errors"

var (
	ErrRecordNotFound      = errors.New("record not found")
	ErrNoRecordDeleted     = errors.New("no record deleted")
	ErrNoRecordUpdated     = errors.New("no record updated")
	ErrNegativeLimit       = errors.New("limit is negative")
	ErrDuplicatedUniqueKey = errors.New("unique key duplicated")
)
