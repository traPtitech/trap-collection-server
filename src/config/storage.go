package config

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import "net/url"

type StorageType int8

const (
	StorageTypeSwift StorageType = iota + 1
	StorageTypeLocal
	StorageTypeS3
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

type StorageS3 interface {
	AccessKeyID() (string, error)
	SecretAccessKey() (string, error)
	Region() (string, error)
	Bucket() (string, error)
	Endpoint() (string, error)
	UsePathStyle() bool
}

type StorageLocal interface {
	Path() (string, error)
}
