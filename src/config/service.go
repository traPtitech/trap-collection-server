package config

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type ServiceV1 interface {
	Administrators() ([]string, error)
	ClientID() (string, error)
	ClientSecret() (string, error)
}
