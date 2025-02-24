package config

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"net/http"
	"net/url"
)

type AuthTraQ interface {
	HTTPClient() (*http.Client, error)
	BaseURL() (*url.URL, error)
}
