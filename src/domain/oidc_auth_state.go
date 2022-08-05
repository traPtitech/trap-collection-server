package domain

import "github.com/traPtitech/trap-collection-server/src/domain/values"

// OIDCAuthState
// OIDCでの認証途中時の状態を表すドメイン。
type OIDCAuthState struct {
	codeChallengeMethod values.OIDCCodeChallengeMethod
	codeVerifier        values.OIDCCodeVerifier
}

func NewOIDCAuthState(
	codeChallengeMethod values.OIDCCodeChallengeMethod,
	codeVerifier values.OIDCCodeVerifier,
) *OIDCAuthState {
	return &OIDCAuthState{
		codeChallengeMethod: codeChallengeMethod,
		codeVerifier:        codeVerifier,
	}
}

func (oas *OIDCAuthState) GetCodeChallengeMethod() values.OIDCCodeChallengeMethod {
	return oas.codeChallengeMethod
}

func (oas *OIDCAuthState) GetCodeVerifier() values.OIDCCodeVerifier {
	return oas.codeVerifier
}
