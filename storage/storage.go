package storage

import (
	"errors"
	"io"
)

// ErrFileNotFound ファイルがないときのエラー
var ErrFileNotFound = errors.New("File Not Found")

// Storage ストレージのインターフェイス
type Storage interface {
	// Save 保存
	Save(filename string, src io.Reader) error
	// Open 開く
	Open(filename string) (io.ReadCloser, error)
	// Delete 削除
	Delete(filename string) error
}
