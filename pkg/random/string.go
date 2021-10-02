package random

import (
	crand "crypto/rand"
	"fmt"
	"unsafe"
)

const (
	alphaNumericLetters    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	alphaNumericLetterMask = 1<<6 - 1
)

/*
	SecureAlphaNumeric
	暗号的に安全なランダム英数字文字列を生成
*/
func SecureAlphaNumeric(length int) (string, error) {
	b := make([]byte, length)
	_, err := crand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to read random bytes: %v", err)
	}

	for i := 0; i < length; {
		idx := int(b[i] & alphaNumericLetterMask)

		if idx < len(alphaNumericLetters) {
			b[i] = alphaNumericLetters[idx]
			i++
		} else {
			// 英数字の文字数62より大きい場合乱数再取得
			_, err := crand.Read(b[i : i+1])
			if err != nil {
				return "", fmt.Errorf("failed to read random bytes: %v", err)
			}
		}
	}

	return *(*string)(unsafe.Pointer(&b)), nil
}
