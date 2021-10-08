package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOIDCSessionIsExpired(t *testing.T) {
	t.Parallel()

	type test struct {
		description string
		expiresAt   time.Time
		expected    bool
	}

	testCases := []test{
		{
			description: "期限前なのでfalse",
			expiresAt:   time.Now().Add(1 * time.Hour),
			expected:    false,
		},
		{
			description: "期限後なのでtrue",
			expiresAt:   time.Now().Add(-1 * time.Hour),
			expected:    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			session := OIDCSession{
				expiresAt: testCase.expiresAt,
			}

			actual := session.IsExpired()
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
