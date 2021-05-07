package values

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"

	"github.com/google/uuid"
)

type (
	GameAssetID string
	GameFileType string
	GameFileMd5 string
	GameURL string
)

const (
	GameFileTypeJar GameFileType = "jar"
	GameFileTypeWindowsExe GameFileType = "exe"
	GameFileTypeMacApp GameFileType = "app"
)

func NewGameAssetID() GameAssetID {
	return GameAssetID(uuid.New().String())
}

func NewGameAssetIDFromString(id string) (GameAssetID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidFormat
	}

	return GameAssetID(id), nil
}

func NewGameFileMd5(reader io.Reader) (GameFileMd5, error) {
	h := md5.New()
	_, err := io.Copy(h, reader)
	if err != nil {
		return "", fmt.Errorf("failed to create md5 hash: %w", err)
	}

	return GameFileMd5(hex.EncodeToString(h.Sum(nil))), nil
}

func NewGameFileMd5FromString(md5 string) (GameFileMd5, error) {
	if len(md5) != 16 {
		return "", ErrInvalidFormat
	}

	return GameFileMd5(md5), nil
}

func NewGameURL(u string) (GameURL, error) {
	// 無印ShowcaseではurlのRFC的にはアウトのunderbarが入りうるのでコメントアウト
	if urlObj, err := 	url.Parse(u); err != nil || !urlObj.IsAbs() {
		return "", ErrInvalidFormat
	}

	return GameURL(u), nil
}
