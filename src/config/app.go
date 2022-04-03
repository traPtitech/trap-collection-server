package config

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type AppStatus int8

const (
	AppStatusProduction AppStatus = iota + 1
	AppStatusDevelopment
)

type App interface {
	Status() (AppStatus, error)
}
