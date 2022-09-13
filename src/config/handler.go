package config

import "net/url"

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type Handler interface {
	Addr() (string, error)
	SessionKey() (string, error)
	SessionSecret() (string, error)
	TraqBaseURL() (*url.URL, error)
}
