package repository

import "errors"

var (
	ErrRecordNotFound     = errors.New("record not found")
	ErrNoRecordDeleted    = errors.New("no record deleted")
	ErrNoRecordUpdated    = errors.New("no record updated")
	ErrNegativeLimit      = errors.New("limit is negative")
	ErrOffsetWithoutLimit = errors.New("there is offset but no limit")
	ErrBadLimitAndOffset  = errors.New("Limit and Offset are invalid")
)
