package encryption

import (
	"reflect"
	"testing"
)

func TestPkcs5Pad(t *testing.T) {
	tests := []struct {
		blockSize int
		unpadded  []byte
		padded    []byte
	}{
		{
			blockSize: 8,
			unpadded:  []byte{0},
			padded:    []byte{0, 7, 7, 7, 7, 7, 7, 7},
		},
		{
			blockSize: 16,
			unpadded:  []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			padded:    []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
	}

	for i, tt := range tests {
		padded := pkcs5Pad(tt.unpadded, tt.blockSize)
		if got, want := padded, tt.padded; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got=%v, want=%v", i, got, want)
		}
	}
}

func TestPkcs5Unpad(t *testing.T) {
	tests := []struct {
		padded   []byte
		unpadded []byte
		errText  string
	}{
		{
			padded:   []byte{0, 7, 7, 7, 7, 7, 7, 7},
			unpadded: []byte{0},
		},
		{
			padded:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			unpadded: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			padded:  []byte{0, 0, 13},
			errText: "invalid PKCS #5 padding",
		},
	}

	for i, tt := range tests {
		unpadded, err := pkcs5Unpad(tt.padded)
		if err != nil {
			if tt.errText == "" {
				t.Errorf("%d: %v", i, err)
				continue
			}
			if got, want := err.Error(), tt.errText; got != want {
				t.Errorf("%d: got=%q want=%q", i, got, want)
				continue
			}
		}
		if got, want := unpadded, tt.unpadded; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got=%v, want=%v", i, got, want)
		}
	}
}
