package config

import "net/url"

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type Handler interface {
	Addr() (string, error)
	SessionKey() (string, error)
	SessionSecret() (string, error)
	TraqBaseURL() (*url.URL, error)
}
