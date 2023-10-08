package values

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameGenreValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		genreName   GameGenreName
		isErr       bool
		expectedErr error
	}

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			genreName: NewGameGenreName("genre"),
			isErr:     false,
		},
		"日本語で32文字でもエラー無し": {
			genreName: NewGameGenreName(strings.Repeat("あ", 32)),
			isErr:     false,
		},
		"空白なのでエラー": {
			genreName:   NewGameGenreName(""),
			isErr:       true,
			expectedErr: ErrGameGenreNameEmpty,
		},
		"32文字以上なのでエラー": {
			genreName:   NewGameGenreName(strings.Repeat("あ", 33)),
			isErr:       true,
			expectedErr: ErrGameGenreNameTooLong,
		},
	}

	for desc, testCase := range testCases {
		t.Run(desc, func(t *testing.T) {
			err := testCase.genreName.Validate()

			if testCase.isErr {
				if testCase.expectedErr == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
