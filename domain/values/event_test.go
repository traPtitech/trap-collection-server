package values

import (
	"errors"
	"testing"
	"testing/quick"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewEventIDFromString(t *testing.T) {
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
		_, err := NewEventIDFromString(test.id)

		if test.err == nil {
			assertion.NoErrorf(err, test.description+"/no error")
		} else {
			assertion.ErrorIs(err, test.err, test.description+"/error")
		}
	}

	err := quick.Check(func(id id) bool {
		_, err := NewEventIDFromString(uuid.UUID(id).String())
		return err == nil
	}, nil)
	if err != nil {
		t.Error("black box test error:", err.Error())
	}
}

func TestNewEventName(t *testing.T) {
	t.Parallel()

	assertion := assert.New(t)

	tests := []struct{
		description string
		name string
		err error
	} {
		{
			description: "32文字なのでエラーなし",
			name: "01234567890123456789012345678901",
			err: nil,
		},
		{
			description: "33文字なのでエラー",
			name: "012345678901234567890123456789012",
			err: ErrTooLong,
		},
	}

	for _, test := range tests {
		_, err := NewEventName(test.name)

		if test.err == nil {
			assertion.NoErrorf(err, test.description+"/no error")
		} else {
			assertion.ErrorIs(err, test.err, test.description+"/error")
		}
	}

	err := quick.Check(func(name string) bool {
		_, err := NewEventName(name)
		return (len(name) >32 && errors.Is(err, ErrTooLong)) || (len(name) <= 32 && err == nil)
	}, nil)
	if err != nil {
		t.Error("black box test error:", err.Error())
	}
}
