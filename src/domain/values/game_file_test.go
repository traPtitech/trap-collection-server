package values

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGameFileHash(t *testing.T) {
	t.Parallel()

	r := strings.NewReader("Beich8gei3pheseen5uuwie7e")
	hash, err := NewGameFileHash(r)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	assert.Equal(t, "a32354ed11d6d65a78cbedac5d55e35f", hash.String())
}
