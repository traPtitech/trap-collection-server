package config

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type AppStatus int8

const (
	AppStatusProduction AppStatus = iota + 1
	AppStatusDevelopment
)

type App interface {
	Status() (AppStatus, error)
	FeatureV2() bool
	FeatureV1Write() bool
}
