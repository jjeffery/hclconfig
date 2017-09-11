package encryption

import (
	"crypto/cipher"
	"crypto/rand"
)

type header []byte

func newHeader(block cipher.Block) header {
	b := make([]byte, block.BlockSize())
	rand.Read(b)

	// Clear the bottom two bits of the first byte for the
	// version number. This will assist if we want to change
	// the format in future.
	b[0] &= 0xfc

	return header(b)
}

func readHeader(block cipher.Block, buf []byte) (header, error) {
	if len(buf) < block.BlockSize() {
		return nil, errInvalidCiphertext
	}
	return header(buf[:block.BlockSize()]), nil
}

func (hdr header) Len() int {
	return len(hdr)
}

func (hdr header) Version() byte {
	return hdr[0] & 0x3
}

func (hdr header) IV(block cipher.Block) []byte {
	iv := make([]byte, block.BlockSize())
	block.Encrypt(iv, hdr)
	return iv
}

func (hdr header) Bytes() []byte {
	return []byte(hdr)
}
