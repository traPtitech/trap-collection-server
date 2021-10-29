package repository

type LockType int

const (
	LockTypeNone LockType = iota
	LockTypeRecord
)
