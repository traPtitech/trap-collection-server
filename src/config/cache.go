package config

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import "time"

type CacheRistretto interface {
	ActiveUsersTTL() (time.Duration, error)
	ActiveSeatsTTL() (time.Duration, error)
}
