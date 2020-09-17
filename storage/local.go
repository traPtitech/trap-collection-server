package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Local 手元での開発時のローカルでのファイルの読み書きのための構造体
type Local struct {
	localDir string
}

// NewLocalStorage LocalStorageのコンストラクタ
func NewLocalStorage(dir string) (*Local, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		return &Local{}, errors.New("dir doesn't exist")
	}
	if !fi.IsDir() {
		return &Local{}, errors.New("dir is not a directory")
	}

	local := &Local{
		localDir: dir,
	}
	return local, nil
}

// Save ファイルの保存
func (l *Local) Save(filename string, src io.Reader) error {
	path := l.getFilePath(filename)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Failed In Creating File: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, src)
	if err != nil {
		return fmt.Errorf("Failed In Copying Data: %w", err)
	}

	return nil
}

// Open ファイルを開く
func (l *Local) Open(filename string) (io.ReadCloser, error) {
	filePath := l.getFilePath(filename)

	r, err := os.Open(filePath)
	if err != nil {
		return nil, ErrFileNotFound
	}

	return r, nil
}

// Delete ファイルの削除
func (l *Local) Delete(filename string) error {
	path := l.getFilePath(filename)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrFileNotFound
	} else if err != nil {
		return fmt.Errorf("Failed In Getting File Info: %w", err)
	}

	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("Failed In Removing File: %w", err)
	}

	return nil
}

func (l *Local) getFilePath(filename string) string {
	path := filepath.Join(l.localDir, filename)

	return path
}
