package values

import (
	"errors"
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

func TestGameFileEntryPointValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		description string
		entryPoint  GameFileEntryPoint
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			entryPoint:  "main.jar",
		},
		{
			description: "空文字列はエラー",
			entryPoint:  "",
			isErr:       true,
			err:         ErrGameFileEntryPointEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := testCase.entryPoint.Validate()

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
