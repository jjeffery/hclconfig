package astcrypt

import (
	"bytes"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/printer"
)

type testEncryptor struct{}

func (t *testEncryptor) EncryptString(cleartext string) (ciphertext string, err error) {
	ciphertext = cleartext
	ciphertext = strings.Replace(ciphertext, "\"", "'", -1)
	return "ciphertext(" + ciphertext + ")", nil
}

func (t *testEncryptor) DecryptString(ciphertext string) (cleartext string, err error) {
	cleartext = strings.TrimSpace(ciphertext)
	cleartext = strings.Replace(cleartext, "'", "\"", -1)
	cleartext = strings.TrimPrefix(cleartext, "ciphertext(")
	cleartext = strings.TrimSuffix(cleartext, ")")
	return cleartext, nil
}

const (
	testdataDir = "testdata"
)

var (
	keywords = []string{"password", "secret"}
	valwords = []string{"password="}
)

func TestEncrypt(t *testing.T) {
	files, err := ioutil.ReadDir(testdataDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		name := f.Name()

		if !strings.HasPrefix(name, "test") {
			continue
		}
		if !strings.HasSuffix(name, "-clear.hcl") {
			continue
		}

		clearBytes, err := ioutil.ReadFile(filepath.Join(testdataDir, name))
		if err != nil {
			log.Fatal(err)
		}

		cipherName := strings.Replace(name, "-clear.", "-cipher.", 1)
		cipherBytes, err := ioutil.ReadFile(filepath.Join(testdataDir, cipherName))
		if err != nil {
			log.Fatal(err)
		}

		node, err := hcl.ParseBytes(clearBytes)
		if err != nil {
			log.Fatal(err)
		}

		if err := Encrypt(node, &testEncryptor{}, keywords, valwords); err != nil {
			log.Fatal(err)
		}

		var buf bytes.Buffer
		prn := printer.Config{
			SpacesWidth: 4,
		}
		prn.Fprint(&buf, node)

		got := buf.String()
		want := string(cipherBytes)
		compareStrings(t, name, got, want)
	}
}

func TestDecrypt(t *testing.T) {
	files, err := ioutil.ReadDir(testdataDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		name := f.Name()

		if !strings.HasPrefix(name, "test") {
			continue
		}
		if !strings.HasSuffix(name, ".hcl") {
			continue
		}

		// testing decrypt is a bit different because we can have multiple
		// ciphertext documents for one cleartext document
		if !strings.Contains(name, "-cipher") {
			continue
		}

		cipherBytes, err := ioutil.ReadFile(filepath.Join(testdataDir, name))
		if err != nil {
			t.Error(name, err)
			continue
		}

		re := regexp.MustCompile(`cipher[0-9-]*`)
		clearName := re.ReplaceAllString(name, "clear")
		clearBytes, err := ioutil.ReadFile(filepath.Join(testdataDir, clearName))
		if err != nil {
			t.Error(name, err)
			continue
		}

		node, err := hcl.ParseBytes(cipherBytes)
		if err != nil {
			t.Error(name, err)
			continue
		}

		if err := Decrypt(node, &testEncryptor{}); err != nil {
			t.Error(name, err)
		}

		var buf bytes.Buffer
		prn := printer.Config{
			SpacesWidth: 4,
		}
		prn.Fprint(&buf, node)

		got := buf.String()
		want := string(clearBytes)
		compareStrings(t, name, got, want)
	}
}

func compareStrings(t *testing.T, name string, got, want string) bool {
	wsRE := regexp.MustCompile(`\s+`)
	got = strings.TrimSpace(got)
	want = strings.TrimSpace(want)
	got = wsRE.ReplaceAllString(got, " ")
	want = wsRE.ReplaceAllString(want, " ")
	if got == want {
		return true
	}

	calcMinLen := func() int {
		n := len(got)
		if n > len(want) {
			n = len(want)
		}
		return n
	}

	// TODO(jpj): fix this to work with non-ASCII characters
	minLen := calcMinLen()

	for skip := 1; skip < minLen; skip++ {
		gotSkip := len(got) - skip
		wantSkip := len(want) - skip
		if got[gotSkip:] != want[wantSkip:] {
			got = got[:gotSkip+1]
			want = want[:wantSkip+1]
			const suffix = " ..."
			got = got + suffix
			want = want + suffix
			break
		}
	}

	minLen = calcMinLen()

	for skip := 1; skip < minLen; skip++ {
		if got[:skip] != want[:skip] {
			got = got[skip-1:]
			want = want[skip-1:]
			const prefix = "..."
			got = prefix + got
			want = prefix + want
			break
		}
	}

	t.Errorf("%s:\n got=%s\nwant=%s", name, got, want)
	return false
}
