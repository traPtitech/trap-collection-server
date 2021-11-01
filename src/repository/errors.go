package repository

import "errors"

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrNoRecordDeleted = errors.New("no record deleted")
	ErrNoRecordUpdated = errors.New("no record updated")
)
