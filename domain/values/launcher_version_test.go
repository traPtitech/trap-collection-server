package values

import (
	"testing"
	"testing/quick"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewLauncherVersionIDFromString(t *testing.T) {
	t.Parallel()

	assertion := assert.New(t)

	tests := []struct{
		description string
		id string
		err error
	}{
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
		_, err := NewLauncherVersionIDFromString(test.id)

		if test.err == nil {
			assertion.NoErrorf(err, test.description+"/no error")
		} else {
			assertion.ErrorIs(err, test.err, test.description+"/error")
		}
	}

	err := quick.Check(func(id id) bool {
		_, err := NewLauncherVersionIDFromString(uuid.UUID(id).String())
		return err == nil
	}, nil)
	if err != nil {
		t.Error("black box test error:", err.Error())
	}
}

func TestNewLauncherVersionName(t *testing.T) {
	t.Parallel()

	assertion := assert.New(t)

	tests := []struct{
		description string
		name string
		err error
	}{
		{
			description: "正しいカレンダーバージョニングなのでエラーなし",
			name: "2006.01.02",
			err: nil,
		},
		{
			description: "正しいカレンダーバージョニング(MODIFIERあり)なのでエラーなし",
			name: "2006.01.02-kodaisai",
			err: nil,
		},
		{
			description: "最後に.があるのは誤りなのでエラー",
			name: "2006.01.02.",
			err: ErrInvalidFormat,
		},
		{
			description: "0埋めでないのは誤りなのでエラー",
			name: "2006.1.2",
			err: ErrInvalidFormat,
		},
		{
			description: "semantic versionは誤りなのでエラー",
			name: "v1.0.0",
			err: ErrInvalidFormat,
		},
	}

	for _, test := range tests {
		_, err := NewLauncherVersionName(test.name)

		if test.err == nil {
			assertion.NoErrorf(err, test.description+"/no error")
		} else {
			assertion.ErrorIs(err, test.err, test.description+"/error")
		}
	}
}

func TestNewQuestionnaireURL(t *testing.T) {
	t.Parallel()

	assertion := assert.New(t)

	tests := []struct{
		description string
		url string
		err error
	}{
		{
			description: "通常のurlなのでエラーなし",
			url: "http://example.com",
			err: nil,
		},
		{
			description: "エンドポイントありでもエラーなし",
			url: "http://example.com/hoge",
			err: nil,
		},
		{
			description: "エンドポイントあり(トレーリングスラッシュあり)でもエラーなし",
			url: "http://example.com/hoge/",
			err: nil,
		},
		{
			description: "httpsでもエラーなし",
			url: "https://example.com",
			err: nil,
		},
		{
			// TODO: :ha:?チェックある意味ないのでなんとかしたい
			description: "スキームが誤っていてもエラーなし",
			url: "htt://example.com/hoge",
			err: nil,
		},
		{
			// TODO: :ha:?チェックある意味ないのでなんとかしたい
			description: "_が入ってもエラーなし",
			url: "https://hackason20_winter_2.trap.show/customtheme-server/gallery",
			err: nil,
		},
		{
			description: "相対パスと解釈されるが、絶対パス飲み許可されるのでエラー",
			url: "hoge",
			err: ErrInvalidFormat,
		},
	}

	for _, test := range tests {
		_, err := NewQuestionnaireURL(test.url)

		if test.err == nil {
			assertion.NoErrorf(err, test.description+"/no error")
		} else {
			assertion.ErrorIs(err, test.err, test.description+"/error")
		}
	}
}
