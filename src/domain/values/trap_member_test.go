package values

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraPMemberNameValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		description    string
		traPMemberName string
		isErr          bool
		err            error
	}

	testCases := []test{
		{
			description:    "正常なtraQ IDなのでエラーなし",
			traPMemberName: "mazrean",
		},
		{
			description:    "空なのでエラー",
			traPMemberName: "",
			isErr:          true,
			err:            ErrTrapMemberNameEmpty,
		},
		{
			description:    "32文字なのでエラーなし",
			traPMemberName: "abcdefghijklmnopqrstuvwxyzabcdef",
		},
		{
			description:    "33文字なのでエラー",
			traPMemberName: "abcdefghijklmnopqrstuvwxyzabcdefg",
			isErr:          true,
			err:            ErrTrapMemberNameTooLong,
		},
		{
			description:    "数字を含んでもエラーなし",
			traPMemberName: "0123456789klmnopqrstuvwxyzabcdef",
		},
		{
			description:    "大文字を含んでもエラーなし",
			traPMemberName: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef",
		},
		{
			description:    "_を含んでもエラーなし",
			traPMemberName: "abcdefghijklmnopqrstuvwxyzabcde_",
		},
		{
			description:    "-を含んでもエラーなし",
			traPMemberName: "abcdefghijklmnopqrstuvwxyzabcde-",
		},
		{
			description:    "マルチバイト文字を含むのでエラー",
			traPMemberName: "abcdefghijklmnopqrstuvwxyzabcdeあ",
			isErr:          true,
			err:            ErrTrapMemberNameInvalidRune,
		},
		{
			description:    "`を含むのでエラー",
			traPMemberName: "abcdefghijklmnopqrstuvwxyzabcde`",
			isErr:          true,
			err:            ErrTrapMemberNameInvalidRune,
		},
		{
			description:    "{を含むのでエラー",
			traPMemberName: "abcdefghijklmnopqrstuvwxyzabcde{",
			isErr:          true,
			err:            ErrTrapMemberNameInvalidRune,
		},
		{
			description:    "[を含むのでエラー",
			traPMemberName: "abcdefghijklmnopqrstuvwxyzabcde[",
			isErr:          true,
			err:            ErrTrapMemberNameInvalidRune,
		},
		{
			description:    "/を含むのでエラー",
			traPMemberName: "abcdefghijklmnopqrstuvwxyzabcde/",
			isErr:          true,
			err:            ErrTrapMemberNameInvalidRune,
		},
		{
			description:    ":を含むのでエラー",
			traPMemberName: "abcdefghijklmnopqrstuvwxyzabcde:",
			isErr:          true,
			err:            ErrTrapMemberNameInvalidRune,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := TraPMemberName(testCase.traPMemberName).Validate()

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
