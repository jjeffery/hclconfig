package encryption

import (
	"crypto/hmac"
	"crypto/sha256"
)

// addHMAC appends the HMAC to the end of the message.
func addHMAC(msg []byte, key []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(msg)
	return append(msg, hash.Sum(nil)...)
}

// checkHMAC strips the HMAC off the end of the message, checks it,
// and returns the message.
func stripHMAC(msg []byte, key []byte) ([]byte, error) {
	if len(msg) < KeyLength {
		return nil, errInvalidCiphertext
	}
	message := msg[:len(msg)-KeyLength]
	messageMAC := msg[len(message):]
	hash := hmac.New(sha256.New, key)
	hash.Write(message)
	expectedMAC := hash.Sum(nil)
	if !hmac.Equal(expectedMAC, messageMAC) {
		return nil, errInvalidCiphertext
	}
	return message, nil
}
