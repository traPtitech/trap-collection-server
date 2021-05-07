package values

import (
	"io"
	"strings"
	"testing"
	"testing/quick"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewGameAssetIDFromString(t *testing.T) {
	t.Parallel()

	assertion := assert.New(t)

	tests := []struct{
		description string
		id string
		err error
	} {
		{
			description: "正しいuuid(大文字)なのでエラーなし",
			id: "BFDF0B4B-0A5A-45E8-A41C-0976D12A115F",
			err: nil,
		},
		{
			description: "正しいuuid(小文字)なのでエラーなし",
			id: "bfdf0b4b-0a5a-45e8-a41c-0976d12a115f",
			err: nil,
		},
		{
			description: "文字数が少ないのでエラー",
			id: "BFDF0B4B-0A5A-45E8-A41C-0976D12A115",
			err: ErrInvalidFormat,
		},
		{
			description: "文字数0なのでエラー",
			id: "",
			err: ErrInvalidFormat,
		},
	}

	for _, test := range tests {
		_, err := NewGameAssetIDFromString(test.id)

		if test.err == nil {
			assertion.NoErrorf(err, test.description+"/no error")
		} else {
			assertion.ErrorIs(test.err, err, test.description+"/error")
		}
	}

	err := quick.Check(func(id id) bool {
		_, err := NewGameIDFromString(uuid.UUID(id).String())
		return err == nil
	}, nil)
	if err != nil {
		t.Error("black box test error:", err.Error())
	}
}

func TestNewGameFileMd5(t *testing.T) {
	t.Parallel()
	
	assertion := assert.New(t)

	tests := []struct{
		description string
		reader io.Reader
		md5 string
		err error
	} {
		{
			description: "アルファベットはハッシュ化できる",
			reader: strings.NewReader("nya nya nya"),
			md5: "27225a004d56c42e01d825b64f2df976",
			err: nil,
		},
		{
			description: "マルチバイト文字もハッシュ化できる",
			reader: strings.NewReader("猫になりたい"),
			md5: "2966bf92448017aa4a4010fd483efe4b",
			err: nil,
		},
	}

	for _, test := range tests {
		md5, err := NewGameFileMd5(test.reader)

		if test.err == nil {
			assertion.NoErrorf(err, test.description+"/no error")
		} else {
			assertion.ErrorIs(test.err, err, test.description+"/error")
		}

		assertion.Equal(test.md5, string(md5), test.description+"/md5")
	}

	err := quick.Check(func(reader string) bool {
		_, err := NewGameFileMd5(strings.NewReader(reader))
		return err == nil
	}, nil)
	if err != nil {
		t.Error("black box test error:", err.Error())
	}
}
