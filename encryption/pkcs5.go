package encryption

import (
	"bytes"
	"errors"
)

var errInvalidPadding = errors.New("invalid PKCS #5 padding")

// pkcs5Pad pads a buffer up to the block size as per PKCS#5
func pkcs5Pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// Unpad unpads a buffer back to its original size as per PKCS #5.
func pkcs5Unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])
	if unpadding > length {
		return nil, errInvalidPadding
	}
	return src[:(length - unpadding)], nil
}
