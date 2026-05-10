package values

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeedbackQuestionTextValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		description  string
		questionText string
		isErr        bool
		err          error
	}

	testCases := []test{
		{
			description:  "英数字なのでエラーなし",
			questionText: "Did you enjoy the game?",
		},
		{
			description:  "英数字256字でもエラーなし",
			questionText: strings.Repeat("a", 256),
		},
		{
			description:  "英数字257字でエラー",
			questionText: strings.Repeat("a", 257),
			isErr:        true,
			err:          ErrFeedbackQuestionTextTooLong,
		},
		{
			description:  "マルチバイト文字256字でもエラーなし",
			questionText: strings.Repeat("あ", 256),
		},
		{
			description:  "マルチバイト文字257字でエラー",
			questionText: strings.Repeat("あ", 257),
			isErr:        true,
			err:          ErrFeedbackQuestionTextTooLong,
		},
		{
			description:  "空文字でエラー",
			questionText: "",
			isErr:        true,
			err:          ErrFeedbackQuestionTextEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			err := FeedbackQuestionText(testCase.questionText).Validate()

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

func TestFeedbackCommentValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		description string
		comment     string
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "空文字でもエラーなし",
			comment:     "",
		},
		{
			description: "通常の文字列でエラーなし",
			comment:     "このゲームは楽しかったです。",
		},
		{
			description: "英数字2000字でもエラーなし",
			comment:     strings.Repeat("a", 2000),
		},
		{
			description: "英数字2001字でエラー",
			comment:     strings.Repeat("a", 2001),
			isErr:       true,
			err:         ErrFeedbackCommentTooLong,
		},
		{
			description: "マルチバイト文字2000字でもエラーなし",
			comment:     strings.Repeat("あ", 2000),
		},
		{
			description: "マルチバイト文字2001字でエラー",
			comment:     strings.Repeat("あ", 2001),
			isErr:       true,
			err:         ErrFeedbackCommentTooLong,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			err := FeedbackComment(testCase.comment).Validate()

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
