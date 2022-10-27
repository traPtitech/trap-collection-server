package v1

import "time"

type CacheRistretto struct{}

func NewCacheRistretto() *CacheRistretto {
	return &CacheRistretto{}
}

func (*CacheRistretto) ActiveUsersTTL() (time.Duration, error) {
	return time.Hour, nil
}

func (*CacheRistretto) ActiveSeatsTTL() (time.Duration, error) {
	return time.Second, nil
}
