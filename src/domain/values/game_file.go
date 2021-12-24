package values

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type (
	GameFileID   uuid.UUID
	GameFileType int
	/*
		GameFileEntryPoint
		ファイルのエントリーポイント。
		ランチャーはFile解凍後にこのパスのファイルを実行する。
	*/
	GameFileEntryPoint string
	/*
		GameFileHash
		ファイルのハッシュ値。
		ランチャーでファイルが壊れていないかの確認に使用する。
	*/
	GameFileHash []byte
)

func NewGameFileID() GameFileID {
	return GameFileID(uuid.New())
}

func NewGameFileIDFromUUID(id uuid.UUID) GameFileID {
	return GameFileID(id)
}

const (
	GameFileTypeJar GameFileType = iota
	GameFileTypeWindows
	GameFileTypeMac
)

func NewGameFileEntryPoint(entryPoint string) GameFileEntryPoint {
	return GameFileEntryPoint(entryPoint)
}

var (
	ErrGameFileEntryPointEmpty = errors.New("entry point must not be empty")
)

func (ep GameFileEntryPoint) Validate() error {
	if len(ep) == 0 {
		return ErrGameFileEntryPointEmpty
	}

	return nil
}

func NewGameFileHash(r io.Reader) (GameFileHash, error) {
	h := md5.New()
	_, err := io.Copy(h, r)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}

	return GameFileHash(h.Sum(nil)), nil
}

func NewGameFileHashFromBytes(hash []byte) GameFileHash {
	return GameFileHash(hash)
}

func (h GameFileHash) String() string {
	return hex.EncodeToString(h)
}
