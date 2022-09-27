package config

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type ServiceV1 interface {
	Administrators() ([]string, error)
	ClientID() (string, error)
	ClientSecret() (string, error)
}

type ServiceV2 interface {
	// ClientID
	// OIDC・OAuth2.0(Authorization Code Flow)のClientIDを取得する
	ClientID() (string, error)
	// ClientSecret
	// OIDC・OAuth2.0(Authorization Code Flow)のClientSecretを取得する
	// traQではSecret関連の機能は未実装なため、基本的に使うことはない
	ClientSecret() (string, error)
}
