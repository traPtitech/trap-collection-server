package io_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/pkg/io"
)

func TestReaderEqual(t *testing.T) {
	testCases := []struct {
		description string
		s1          string
		s2          string
		isEqual     bool
		isErr       bool
	}{
		{
			description: "s1とs2が同じ",
			s1:          "abc",
			s2:          "abc",
			isEqual:     true,
		},
		{
			description: "s1とs2が異なる",
			s1:          "abc",
			s2:          "def",
			isEqual:     false,
		},
		{
			description: "s1の方が長い",
			s1:          "abcdef",
			s2:          "abc",
			isEqual:     false,
		},
		{
			description: "s2の方が長い",
			s1:          "abc",
			s2:          "abcdef",
			isEqual:     false,
		},
		{
			description: "s1が空文字列",
			s1:          "",
			s2:          "abc",
			isEqual:     false,
		},
		{
			description: "s2が空文字列",
			s1:          "abc",
			s2:          "",
			isEqual:     false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			r1 := strings.NewReader(testCase.s1)
			r2 := strings.NewReader(testCase.s2)

			isEqual, err := io.ReaderEqual(r1, r2)

			if testCase.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, testCase.isEqual, isEqual)
		})
	}
}
