package config

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock -typed

type Migration interface {
	EmptyDB() (bool, error)
	Baseline() (string, error)
}
