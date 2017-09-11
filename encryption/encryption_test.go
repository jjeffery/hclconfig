package encryption

import "testing"

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
