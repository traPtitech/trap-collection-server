package v1

import (
	"errors"
	"os"
	"strings"
)

type ServiceV1 struct{}

func NewServiceV1() *ServiceV1 {
	return &ServiceV1{}
}

func (*ServiceV1) Administrators() ([]string, error) {
	strAdministrators, ok := os.LookupEnv(envKeyAdministrators)
	if !ok {
		return nil, errors.New("ADMINISTRATORS is not set")
	}

	administrators := strings.Split(strings.TrimSpace(strAdministrators), ",")

	return administrators, nil
}

func (*ServiceV1) ClientID() (string, error) {
	clientID, ok := os.LookupEnv(envKeyClientID)
	if !ok {
		return "", errors.New("ENV CLIENT_ID IS NULL")
	}

	return clientID, nil
}

func (*ServiceV1) ClientSecret() (string, error) {
	clientSecret, ok := os.LookupEnv(envKeyClientSecret)
	if !ok {
		return "", errors.New("ENV CLIENT_SECRET IS NULL")
	}

	return clientSecret, nil
}
