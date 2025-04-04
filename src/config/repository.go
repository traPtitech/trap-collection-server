package config

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type RepositoryGorm2 interface {
	User() (string, error)
	Password() (string, error)
	Host() (string, error)
	Port() (int, error)
	Database() (string, error)
}
