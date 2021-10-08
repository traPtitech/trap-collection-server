package values

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/traPtitech/trap-collection-server/pkg/random"
)

type (
	OIDCClientID string

	OIDCAuthorizationCode   string
	OIDCCodeChallengeMethod int
	OIDCCodeChallenge       string
	OIDCCodeVerifier        string

	OIDCAccessToken string
)

const (
	OIDCCodeChallengeMethodSha256 OIDCCodeChallengeMethod = iota
)

func NewOIDCClientID(clientID string) OIDCClientID {
	return OIDCClientID(clientID)
}

func NewOIDCAuthorizationCode(code string) OIDCAuthorizationCode {
	return OIDCAuthorizationCode(code)
}

func NewOIDCCodeVerifier() (OIDCCodeVerifier, error) {
	randStr, err := random.SecureAlphaNumeric(64)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	return OIDCCodeVerifier(randStr), nil
}

func NewOIDCCodeVerifierFromString(randStr string) OIDCCodeVerifier {
	return OIDCCodeVerifier(randStr)
}

func (ocd OIDCCodeVerifier) GetCodeChallenge(method OIDCCodeChallengeMethod) (OIDCCodeChallenge, error) {
	switch method {
	case OIDCCodeChallengeMethodSha256:
		//SHA256でハッシュ化
		hash := sha256.New()
		_, err := strings.NewReader(string(ocd)).WriteTo(hash)
		if err != nil {
			return "", fmt.Errorf("failed to write to hash: %w", err)
		}

		//BASE64URLエンコード
		codeChallengeBuilder := strings.Builder{}
		encoder := base64.NewEncoder(base64.RawURLEncoding, &codeChallengeBuilder)
		_, err = encoder.Write(hash.Sum(nil))
		if err != nil {
			return "", fmt.Errorf("failed to encode hash: %w", err)
		}
		encoder.Close()

		return OIDCCodeChallenge(codeChallengeBuilder.String()), nil
	}

	return "", fmt.Errorf("unsupported code challenge method: %d", method)
}

func NewAccessToken(accessToken string) OIDCAccessToken {
	return OIDCAccessToken(accessToken)
}
