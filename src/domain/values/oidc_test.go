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
