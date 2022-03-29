package v1

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type RepositoryGorm2 struct{}

func NewRepositoryGorm2() *RepositoryGorm2 {
	return &RepositoryGorm2{}
}

func (*RepositoryGorm2) User() (string, error) {
	user, ok := os.LookupEnv(envKeyDBUserName)
	if !ok {
		return "", errors.New("DB_USERNAME is not set")
	}

	return user, nil
}

func (*RepositoryGorm2) Password() (string, error) {
	password, ok := os.LookupEnv(envKeyDBPassword)
	if !ok {
		return "", errors.New("DB_PASSWORD is not set")
	}

	return password, nil
}

func (*RepositoryGorm2) Host() (string, error) {
	host, ok := os.LookupEnv(envKeyDBHostName)
	if !ok {
		return "", errors.New("DB_HOSTNAME is not set")
	}

	return host, nil
}

func (*RepositoryGorm2) Port() (int, error) {
	strPort, ok := os.LookupEnv(envKeyDBPort)
	if !ok {
		return 0, errors.New("DB_PORT is not set")
	}

	port, err := strconv.Atoi(strPort)
	if err != nil {
		return 0, fmt.Errorf("DB_PORT is not a number: %s", strPort)
	}

	return port, nil
}

func (*RepositoryGorm2) Database() (string, error) {
	database, ok := os.LookupEnv(envKeyDBDatabase)
	if !ok {
		return "", errors.New("DB_DATABASE is not set")
	}

	return database, nil
}
