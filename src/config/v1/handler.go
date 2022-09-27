package v1

import (
	"errors"
	"fmt"
	"net/url"
	"os"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (*Handler) Addr() (string, error) {
	port, ok := os.LookupEnv(envKeyPort)
	if !ok {
		return "", errors.New("PORT is not set")
	}

	return port, nil
}

func (*Handler) SessionKey() (string, error) {
	return "sessions", nil
}

func (*Handler) SessionSecret() (string, error) {
	secret, ok := os.LookupEnv(envKeySessionSecret)
	if !ok {
		return "", errors.New("SESSION_SECRET is not set")
	}

	return secret, nil
}

func (*Handler) TraqBaseURL() (*url.URL, error) {
	traQBaseURL, err := url.Parse("https://q.trap.jp/api/v3")
	if err != nil {
		return nil, fmt.Errorf("failed to parse traQBaseURL: %w", err)
	}

	return traQBaseURL, nil
}
