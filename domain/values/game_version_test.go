package values

import (
	"testing"
	"testing/quick"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewGameVersionIDFromString(t *testing.T) {
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
		_, err := NewGameVersionIDFromString(test.id)

		if test.err == nil {
			assertion.NoErrorf(err, test.description+"/no error")
		} else {
			assertion.ErrorIs(err, test.err, test.description+"/error")
		}
	}

	err := quick.Check(func(id id) bool {
		_, err := NewGameVersionIDFromString(uuid.UUID(id).String())
		return err == nil
	}, nil)
	if err != nil {
		t.Error("black box test error:", err.Error())
	}
}

func TestNewGameVersionName(t *testing.T) {
	t.Parallel()

	assertion := assert.New(t)

	tests := []struct{
		description string
		name string
		err error
	} {
		{
			description: "正しいsemantic versionなのでエラーなし",
			name: "v1.0.0",
			err: nil,
		},
		{
			description: "PATCHなしでも正しいのでエラーなし",
			name: "v1.0",
			err: nil,
		},
		{
			description: "MINORなしでも正しいのでエラーなし",
			name: "v1",
			err: nil,
		},
		{
			description: "PRERELEASEありでも正しいのでエラーなし",
			name: "v1.0.0-rc",
			err: nil,
		},
		{
			description: "BUILDありでも正しいのでエラーなし",
			name: "v1.0.0+dev",
			err: nil,
		},
		{
			description: "PRERELEASEあり、BUILDありでも正しいのでエラーなし",
			name: "v1.0.0-rc+dev",
			err: nil,
		},
		{
			description: "PATCHなしでのPRERELEASEありは誤りなのでエラー",
			name: "v1.0-rc",
			err: ErrInvalidFormat,
		},
		{
			description: "PATCHなしでのBUILDありは誤りなのでエラー",
			name: "v1.0+dev",
			err: ErrInvalidFormat,
		},
		{
			description: "誤りなのでエラーなし",
			name: "v1.0.",
			err: ErrInvalidFormat,
		},
		{
			description: "文字数0なのでエラー",
			name: "",
			err: ErrInvalidFormat,
		},
	}

	for _, test := range tests {
		_, err := NewGameVersionName(test.name)

		if test.err == nil {
			assertion.NoErrorf(err, test.description+"/no error")
		} else {
			assertion.ErrorIs(err, test.err, test.description+"/error")
		}
	}
}
