package v1

import "bytes"

// NopCloser テストでbytes.Readerをmultipart.Fileに対応するための構造体
type NopCloser struct {
	*bytes.Reader
}

func (r *NopCloser) Close() error {
	return nil
}
