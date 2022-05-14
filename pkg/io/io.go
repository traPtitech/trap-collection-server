package io

import (
	"fmt"
	"io"
)

func ReaderEqual(r1 io.Reader, r2 io.Reader) (bool, error) {
	buf1 := make([]byte, 1024)
	buf2 := make([]byte, 1024)

	i := 0
	isEOF1 := false
	isEOF2 := false
	for ; ; i++ {
		n1, err := r1.Read(buf1)
		if err != nil {
			if err == io.EOF {
				isEOF1 = true
			} else {
				return false, fmt.Errorf("failed to read r1: %w", err)
			}
		}

		n2, err := r2.Read(buf2)
		if err != nil {
			if err == io.EOF {
				isEOF2 = true
			} else {
				return false, fmt.Errorf("failed to read r2: %w", err)
			}
		}

		if n1 != n2 {
			return false, nil
		}

		for i := 0; i < n1; i++ {
			if buf1[i] != buf2[i] {
				return false, nil
			}
		}

		if isEOF1 {
			return isEOF2, nil
		}
		if isEOF2 {
			return false, nil
		}
	}
}
