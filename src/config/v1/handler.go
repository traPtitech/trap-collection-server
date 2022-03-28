package v1

import (
	"errors"
	"fmt"
	"net/url"
	"os"
)

type HandlerV1 struct{}

func NewHandlerV1() *HandlerV1 {
	return &HandlerV1{}
}

func (*HandlerV1) Addr() (string, error) {
	port, ok := os.LookupEnv(envKeyPort)
	if !ok {
		return "", errors.New("PORT is not set")
	}

	return port, nil
}

func (*HandlerV1) SessionKey() (string, error) {
	return "sessions", nil
}

func (*HandlerV1) SessionSecret() (string, error) {
	secret, ok := os.LookupEnv(envKeySessionSecret)
	if !ok {
		return "", errors.New("SESSION_SECRET is not set")
	}

	return secret, nil
}

func (*HandlerV1) TraqBaseURL() (*url.URL, error) {
	traQBaseURL, err := url.Parse("https://q.trap.jp/api/v3")
	if err != nil {
		return nil, fmt.Errorf("failed to parse traQBaseURL: %w", err)
	}

	return traQBaseURL, nil
}
