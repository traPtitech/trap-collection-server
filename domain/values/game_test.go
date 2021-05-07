package values

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type id uuid.UUID
func (*id) Generate(rand *rand.Rand, size int) reflect.Value {
	uuidObj, _ := uuid.NewRandom()

	return reflect.ValueOf(id(uuidObj))
}

func TestNewGameIDFromString(t *testing.T) {
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
		_, err := NewGameIDFromString(test.id)

		if test.err == nil {
			assertion.NoErrorf(err, test.description+"/no error")
		} else {
			assertion.ErrorIs(test.err, err, test.description+"/error")
		}
	}

	err := quick.Check(func(id id) bool {
		_, err := NewGameIDFromString(uuid.UUID(id).String())
		t.Log(id)
		return err == nil
	}, nil)
	if err != nil {
		t.Error("black box test error:", err.Error())
	}
}
