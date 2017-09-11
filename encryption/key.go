// Package encryption performs encryption using AES256 CBC + HMAC SHA256.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"

	"github.com/jjeffery/errors"
)

const (
	// KeyLength is the expected key length, in bytes.
	KeyLength = 256 / 8
)

var (
	errInvalidCiphertext = errors.New("invalid ciphertext")
)

// Key is a 256 bit encryption key.
type Key []byte

func (key Key) checkLength() error {
	if got, want := len(key), KeyLength; got != want {
		return errors.New("incorrect key length").With(
			"got", got,
			"want", want,
		)
	}
	return nil
}

// Encrypt cleartext bytes into a base64 ciphertext string
func (key Key) Encrypt(cleartext []byte) (ciphertext string, err error) {
	if err = key.checkLength(); err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "cannot encrypt")
	}

	// The header contains a nonce for generating the IV.
	header := newHeader(block)

	// Pad the clear text and encrypt.
	paddedMsg := pkcs5Pad(cleartext, block.BlockSize())
	iv := header.IV(block)
	cbcEncrypter := cipher.NewCBCEncrypter(block, iv)
	cbcEncrypter.CryptBlocks(paddedMsg, paddedMsg)

	// allocate enough room for the header, cipher text and HMAC.
	b := make([]byte, 0, header.Len()+len(paddedMsg)+KeyLength)
	b = append(b, header.Bytes()...)
	b = append(b, paddedMsg...)
	b = addHMAC(b, key)

	ciphertext = base64.StdEncoding.EncodeToString(b)
	return ciphertext, nil
}

// Decrypt the ciphertext.
func (key Key) Decrypt(ciphertext string) (cleartext []byte, err error) {
	if err = key.checkLength(); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	replacer := strings.NewReplacer(
		"\n", "",
		"\r", "",
		"\t", "",
		" ", "",
	)
	ciphertext = replacer.Replace(ciphertext)

	msg, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.Wrap(err, "cannot decrypt")
	}
	header, err := readHeader(block, msg)
	if err != nil {
		return nil, err
	}

	if header.Version() != 0 {
		return nil, errors.New("invalid ciphertext version")
	}
	msg, err = stripHMAC(msg, key)
	if err != nil {
		return nil, err
	}
	msg = msg[header.Len():]
	if len(msg) == 0 || len(msg)%block.BlockSize() != 0 {
		// the cipher text length is not a multiple of the
		// AES cipher block size, so the input is invalid
		return nil, errInvalidCiphertext
	}
	iv := header.IV(block)
	cbcDecrypter := cipher.NewCBCDecrypter(block, iv)
	cbcDecrypter.CryptBlocks(msg, msg)
	msg, err = pkcs5Unpad(msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// EncryptString encrypts a string value.
func (key Key) EncryptString(cleartext string) (ciphertext string, err error) {
	return key.Encrypt([]byte(cleartext))
}

// DecryptString decrypts the ciphertext into a string value.
func (key Key) DecryptString(ciphertext string) (cleartext string, err error) {
	b, err := key.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
