package values

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameVersionNameValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		description string
		versionName string
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "正しいSemantic Versionなのでエラーなし",
			versionName: "v1.2.3",
			isErr:       false,
		},
		{
			description: "バージョン名が空文字列なのでエラー",
			versionName: "",
			isErr:       true,
			err:         ErrGameVersionNameInvalidSemanticVersion,
		},
		{
			description: "メジャーバージョンのみでもエラーなし",
			versionName: "v1",
			isErr:       false,
		},
		{
			description: "メジャーバージョンとマイナーバージョンのみでもエラーなし",
			versionName: "v1.2",
			isErr:       false,
		},
		{
			description: "PRERELEASEありでもエラーなし",
			versionName: "v1.2.3-alpha",
			isErr:       false,
		},
		{
			description: "BUILDありでもエラーなし",
			versionName: "v1.2.3+build",
			isErr:       false,
		},
		{
			description: "PRERELEASEとBUILDありでもエラーなし",
			versionName: "v1.2.3-alpha+build",
			isErr:       false,
		},
		{
			description: "数字以外の文字列を含むバージョン名なのでエラー",
			versionName: "v1.a.3",
			isErr:       true,
			err:         ErrGameVersionNameInvalidSemanticVersion,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := GameVersionName(testCase.versionName).Validate()

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
