package v1

import (
	"fmt"
	"net/http"
	"net/url"
)

type AuthTraQ struct{}

func NewAuthTraQ() *AuthTraQ {
	return &AuthTraQ{}
}

func (*AuthTraQ) HTTPClient() (*http.Client, error) {
	return http.DefaultClient, nil
}

func (*AuthTraQ) BaseURL() (*url.URL, error) {
	traQBaseURL, err := url.Parse("https://q.trap.jp/api/v3")
	if err != nil {
		return nil, fmt.Errorf("failed to parse traQBaseURL: %w", err)
	}

	return traQBaseURL, nil
}
