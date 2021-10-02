package values

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLauncherVersionNameValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		description string
		versionName string
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "英数字なのでエラーなし",
			versionName: "abcdefg",
			isErr:       false,
		},
		{
			description: "英数字32字でもエラーなし",
			versionName: "abcdefghijklmnopqrstuvwxyz012345",
			isErr:       false,
		},
		{
			description: "英数字33字でエラー",
			versionName: "abcdefghijklmnopqrstuvwxyz0123456",
			isErr:       true,
			err:         ErrLauncherVersionNameTooLong,
		},
		{
			description: "マルチバイト文字32字でもエラーなし",
			versionName: "あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみ",
			isErr:       false,
		},
		{
			description: "マルチバイト文字33字でエラー",
			versionName: "あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむ",
			isErr:       true,
			err:         ErrLauncherVersionNameTooLong,
		},
		{
			description: "空文字でエラー",
			versionName: "",
			isErr:       true,
			err:         ErrLauncherVersionNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := LauncherVersionName(testCase.versionName).Validate()

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
