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

func TestLauncherUserProductKey(t *testing.T) {
	t.Parallel()

	type test struct {
		description string
		productKey  string
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "英語小文字25文字の5文字ごとでの-区切りなのでエラーなし",
			productKey:  "bbcde-fghij-klmno-pqrst-uvwxy",
			isErr:       false,
		},
		{
			description: "aを含んでもエラーなし",
			productKey:  "abcde-fghij-klmno-pqrst-uvwxy",
			isErr:       false,
		},
		{
			description: "zを含んでもエラーなし",
			productKey:  "abcde-fghij-klmno-pqrst-uvwxz",
			isErr:       false,
		},
		{
			description: "`を含むのでエラー",
			productKey:  "`bcde-fghij-klmno-pqrst-uvwxy",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
		{
			description: "{を含むのでエラー",
			productKey:  "abcde-fghij-klmno-pqrst-uvwx}",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
		{
			description: "英語大文字を含んでもエラーなし",
			productKey:  "abcde-fghij-klmno-pQrst-uvwxy",
			isErr:       false,
		},
		{
			description: "Zを含んでもエラーなし",
			productKey:  "abcde-fghij-klmno-pqrst-uvwxZ",
			isErr:       false,
		},
		{
			description: "Aを含んでもエラーなし",
			productKey:  "abcde-fghij-klmno-pqrst-uvwxA",
			isErr:       false,
		},
		{
			description: "[を含むのでエラー",
			productKey:  "abcde-fghij-klmno-pqrst-uvwx[",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
		{
			description: "@を含むのでエラー",
			productKey:  "@bcde-fghij-klmno-pqrst-uvwxy",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
		{
			description: "英語大文字のみでもエラーなし",
			productKey:  "ABCDE-FGHIJ-KLMNO-PQRST-UVWXY",
			isErr:       false,
		},
		{
			description: "数字を含んでもエラーなし",
			productKey:  "abcde-fghij-klmno-pqrst-uvwx1",
			isErr:       false,
		},
		{
			description: "0を含んでもエラーなし",
			productKey:  "abcde-fghij-klmno-pqrst-uvwx0",
			isErr:       false,
		},
		{
			description: "9を含んでもエラーなし",
			productKey:  "abcde-fghij-klmno-pqrst-uvwx9",
			isErr:       false,
		},
		{
			description: "/を含むのでエラー",
			productKey:  "abcde-fghij-klmno-pqrst-uvwx/",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
		{
			description: ":を含むのでエラー",
			productKey:  "abcde-fghij-klmno-pqrst-uvwx:",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
		{
			description: "数字のみでもエラーなし",
			productKey:  "12345-67890-12345-67890-12345",
			isErr:       false,
		},
		{
			description: "文字数が30文字以上なのでエラー",
			productKey:  "abcde-fghij-klmno-pqrst-uvwxyz",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidLength,
		},
		{
			description: "文字数が28文字以下なのでエラー",
			productKey:  "abcde-fghij-klmno-pqrst-uvwx",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidLength,
		},
		{
			description: "6番目が-でないのでエラー",
			productKey:  "abcdezfghij-klmno-pqrst-uvwxy",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
		{
			description: "12番目が-でないのでエラー",
			productKey:  "abcde-fghijzklmno-pqrst-uvwxy",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
		{
			description: "18番目が-でないのでエラー",
			productKey:  "abcde-fghij-klmnozpqrst-uvwxy",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
		{
			description: "24番目が-でないのでエラー",
			productKey:  "abcde-fghij-klmno-pqrstzuvwxy",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
		{
			description: "6,12,18,24番目以外に-があるのでエラー",
			productKey:  "abcd--fghij-klmno-pqrst-uvwxy",
			isErr:       true,
			err:         ErrLauncherUserProductKeyInvalidRune,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := LauncherUserProductKey(testCase.productKey).Validate()

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

func TestLauncherSessionAccessTokenValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		description string
		accessToken string
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "英語小文字64文字なのでエラーなし",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnopq",
			isErr:       false,
		},
		{
			description: "aを含んでもエラーなし",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnopa",
			isErr:       false,
		},
		{
			description: "zを含んでもエラーなし",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnopz",
			isErr:       false,
		},
		{
			description: "`を含むのでエラー",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnop`",
			isErr:       true,
			err:         ErrLauncherSessionAccessTokenInvalidRune,
		},
		{
			description: "{を含むのでエラー",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnop{",
			isErr:       true,
			err:         ErrLauncherSessionAccessTokenInvalidRune,
		},
		{
			description: "英語大文字を含んでもエラーなし",
			accessToken: "Bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnopq",
			isErr:       false,
		},
		{
			description: "Zを含んでもエラーなし",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnopZ",
			isErr:       false,
		},
		{
			description: "Aを含んでもエラーなし",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnopA",
			isErr:       false,
		},
		{
			description: "[を含むのでエラー",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnop[",
			isErr:       true,
			err:         ErrLauncherSessionAccessTokenInvalidRune,
		},
		{
			description: "@を含むのでエラー",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnop@",
			isErr:       true,
			err:         ErrLauncherSessionAccessTokenInvalidRune,
		},
		{
			description: "英語大文字のみでもエラーなし",
			accessToken: "BCDEFGHIJKLMNOPQRSTUVWXYBCDEFGHIJKLMNOPQRSTUVWXYBCDEFGHIJKLMNOPQ",
			isErr:       false,
		},
		{
			description: "数字を含んでもエラーなし",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnop1",
			isErr:       false,
		},
		{
			description: "0を含んでもエラーなし",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnop0",
			isErr:       false,
		},
		{
			description: "9を含んでもエラーなし",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnop9",
			isErr:       false,
		},
		{
			description: "/を含むのでエラー",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnop/",
			isErr:       true,
			err:         ErrLauncherSessionAccessTokenInvalidRune,
		},
		{
			description: ":を含むのでエラー",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnop:",
			isErr:       true,
			err:         ErrLauncherSessionAccessTokenInvalidRune,
		},
		{
			description: "数字のみでもエラーなし",
			accessToken: "1234567891234567891234567891234567891234567891234567891234567891",
			isErr:       false,
		},
		{
			description: "文字数が65文字以上なのでエラー",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnopqr",
			isErr:       true,
			err:         ErrLauncherSessionAccessTokenInvalidLength,
		},
		{
			description: "文字数が63文字以下なのでエラー",
			accessToken: "bcdefghijklmnopqrstuvwxybcdefghijklmnopqrstuvwxybcdefghijklmnop",
			isErr:       true,
			err:         ErrLauncherSessionAccessTokenInvalidLength,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := LauncherSessionAccessToken(testCase.accessToken).Validate()

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
