package encryption

import (
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		key        Key
		cleartext  string
		ciphertext string
	}{
		{
			key: Key{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
				0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			},
			cleartext: "hello world",
		},
		{
			key: Key{
				0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
				0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
				0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			},
			cleartext: "now is the time for all good men to come to the aid of the party",
		},
	}

	for i, tt := range tests {
		key := tt.key

		ciphertext, err := key.EncryptString(tt.cleartext)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		// insert some spaces
		for _, pos := range []int{32, 66, 100, 134} {
			if len(ciphertext) > pos {
				ciphertext = ciphertext[:pos] + "\n " + ciphertext[pos:]
			}
		}
		ciphertext = " " + ciphertext + " "

		cleartext, err := key.DecryptString(ciphertext)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		if got, want := cleartext, tt.cleartext; got != want {
			t.Errorf("%d: got=%s, want=%s", i, got, want)
		}

		t.Logf("%d: %q\n%s", i, cleartext, ciphertext)
	}
}

// TestDecrypt tests decrypting strings and tests backward
// compatibility in the event that changes are made to the
// decryption algorithm.
func TestDecrypt(t *testing.T) {
	tests := []struct {
		key        Key
		cipherText string
		clearText  string
		errText    string
	}{
		{
			// Version 0
			key: Key{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
				0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			},
			cipherText: "oEXNDp3tIpcLoSBpLpD3OKCO6YuQaRL6Odp6J" +
				"Aye1pQwXR8yQf9QI8S+Tsy1iTL7pk+En5z6UKjnIE7LRh" +
				"uhbw==",
			clearText: "hello world",
		},
		{
			key: Key{
				0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
				0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
				0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			},
			cipherText: "yNccukj+CPCYZHm1m2qgCM9kv+tNo3aH6+yIHU" +
				"El4LujwghrJQQ/2eq+pevI4DV9ywqVoQJMqV7ElToccyNA" +
				"ce6tr9ofug9n6Bk/toO+fkJQZUpNh9ROLcWAVtWv3gDt0J" +
				"d79kDDJ/H5DSusZDwm3b+brViXqad2SRu821ju9fU=",
			clearText: "now is the time for all good men to come to the aid of the party",
		},
	}

	for i, tt := range tests {
		clearText, err := tt.key.DecryptString(tt.cipherText)
		if err != nil {
			if tt.errText == "" {
				t.Errorf("%d: %v", i, err)
				continue
			}
			if got, want := err.Error(), tt.errText; got != want {
				t.Errorf("%d: got=%q want=%q", i, got, want)
				continue
			}
			continue
		}

		if got, want := clearText, tt.clearText; got != want {
			t.Errorf("%d: got=%q want=%q", i, got, want)
			continue
		}
	}
}

func TestEncryptErrors(t *testing.T) {
	tests := []struct {
		key       Key
		cleartext string
		errText   string
	}{
		{
			key:     Key{1, 2},
			errText: "incorrect key length",
		},
	}

	for i, tt := range tests {
		_, err := tt.key.EncryptString(tt.cleartext)
		if err == nil {
			t.Errorf("%d: expected error", i)
			continue
		}
		if errText := err.Error(); !strings.Contains(errText, tt.errText) {
			t.Errorf("%d: error=%q, expected it to contain %q", i, errText, tt.errText)
			continue
		}
	}
}

func TestDecryptErrors(t *testing.T) {
	tests := []struct {
		key        Key
		ciphertext string
		errText    string
	}{
		{
			key:     Key{1, 2},
			errText: "incorrect key length",
		},
		{
			key: Key{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			},
			ciphertext: "a!@$%#$@#%#$%#$$%",
			errText:    "illegal base64 data",
		},
		{
			key: Key{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			},
			ciphertext: "abcd",
			errText:    "invalid ciphertext",
		},
		{
			key: Key{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			},
			// ciphertext contents is smaller than the key length
			ciphertext: "zbcda1c3v47f/wfghuwlbuxns9ndlrundlehjkj=",
			errText:    "invalid ciphertext version",
		},
		{
			key: Key{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			},
			ciphertext: "oEXNDp3tIpcLoSBpLpD3OKCO6YuQaRL6Odp6Joo=",
			errText:    "invalid ciphertext",
		},
		{
			key: Key{
				0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
				0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
				0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			},
			// This would be valid ciphertext, except at the end
			// the text "U=" was changed to "Z=".
			// Test case is where the HMAC does not match.
			ciphertext: "yNccukj+CPCYZHm1m2qgCM9kv+tNo3aH6+yIHU" +
				"El4LujwghrJQQ/2eq+pevI4DV9ywqVoQJMqV7ElToccyNA" +
				"ce6tr9ofug9n6Bk/toO+fkJQZUpNh9ROLcWAVtWv3gDt0J" +
				"d79kDDJ/H5DSusZDwm3b+brViXqad2SRu821ju9fZ=",
			errText: "invalid ciphertext",
		},
	}

	for i, tt := range tests {
		_, err := tt.key.DecryptString(tt.ciphertext)
		if err == nil {
			t.Errorf("%d: expected error", i)
			continue
		}
		if errText := err.Error(); !strings.Contains(errText, tt.errText) {
			t.Errorf("%d: error=%q, expected it to contain %q", i, errText, tt.errText)
			continue
		}
	}
}
