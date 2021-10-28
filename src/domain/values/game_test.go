package values

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameNameValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		description string
		gameName    string
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "英数字なのでエラーなし",
			gameName:    "abcdefg",
			isErr:       false,
		},
		{
			description: "英数字32字でもエラーなし",
			gameName:    "abcdefghijklmnopqrstuvwxyz012345",
			isErr:       false,
		},
		{
			description: "英数字33字でエラー",
			gameName:    "abcdefghijklmnopqrstuvwxyz0123456",
			isErr:       true,
			err:         ErrGameNameTooLong,
		},
		{
			description: "マルチバイト文字32字でもエラーなし",
			gameName:    "あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみ",
			isErr:       false,
		},
		{
			description: "マルチバイト文字33字でエラー",
			gameName:    "あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむ",
			isErr:       true,
			err:         ErrGameNameTooLong,
		},
		{
			description: "空文字でエラー",
			gameName:    "",
			isErr:       true,
			err:         ErrGameNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := GameName(testCase.gameName).Validate()

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
