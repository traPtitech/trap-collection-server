package cache

import "errors"

var (
	ErrCacheMiss = errors.New("cache: key not found")
)
