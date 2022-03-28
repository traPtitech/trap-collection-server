package config

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import "net/url"

type StorageType int8

const (
	StorageTypeSwift StorageType = iota + 1
	StorageTypeLocal
)

type Storage interface {
	Type() (StorageType, error)
}

type StorageSwift interface {
	AuthURL() (*url.URL, error)
	UserName() (string, error)
	Password() (string, error)
	TenantID() (string, error)
	TenantName() (string, error)
	Container() (string, error)
	TmpURLKey() (string, error)
}

type StorageLocal interface {
	Path() (string, error)
}
