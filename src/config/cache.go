package config

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import "time"

type CacheRistretto interface {
	ActiveUsersTTL() (time.Duration, error)
}
