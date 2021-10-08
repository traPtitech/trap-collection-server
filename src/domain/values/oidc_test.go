package values

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOIDCCodeVerifier(t *testing.T) {
	t.Parallel()

	loopNum := 100
	accessTokenRegexp, err := regexp.Compile("^[0-9a-zA-Z]{64}$")
	if err != nil {
		t.Errorf("failed to compile product key regexp: %v", err)
	}

	for i := 0; i < loopNum; i++ {
		codeVerifier, err := NewOIDCCodeVerifier()
		assert.NoError(t, err)

		assert.Regexp(t, accessTokenRegexp, codeVerifier)
	}
}

func TestGetCodeChallenge(t *testing.T) {
	t.Parallel()

	type test struct {
		description           string
		codeVerifier          string
		hashMethod            OIDCCodeChallengeMethod
		expectedCodeChallenge string
	}

	testCases := []test{
		{
			description:           "sha256",
			codeVerifier:          "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			hashMethod:            OIDCCodeChallengeMethodSha256,
			expectedCodeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			codeChallenge, err := OIDCCodeVerifier(testCase.codeVerifier).GetCodeChallenge(testCase.hashMethod)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedCodeChallenge, string(codeChallenge))
		})
	}
}
