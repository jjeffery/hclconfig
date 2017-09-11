// Package astutil helps with parsing HCL and manipulating the AST.
package astutil

import (
	"strings"
	"unicode"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/jjeffery/errors"
)

// Decrypter is an interface for decrypting text.
type Decrypter interface {
	DecryptString(ciphertext string) (cleartext string, err error)
}

// Decrypt decrypts the HCL or JSON document in the abstract
// syntax tree. It does this by walking the tree looking for
// ciphertext values of the form
//	key {
//		ciphertext = "<encrypted-data>"
//	}
// or the equivalent JSON
//  "key": { "ciphertext": "<encrypted-data>" }
// These encrypted values are decrypted using the encryption data
// key information in the configuration file and converted into
// values of the form
//  key = "<decrypted-data>"
// The decrypter is used to decrypt ciphertext. If the decrypter
// is nil this function will return success only if there is nothing
// in the AST to decrypt.
func Decrypt(node ast.Node, decrypter Decrypter) error {
	walker := decryptionWalker{
		decrypter: decrypter,
	}

	ast.Walk(node, walker.Walk)
	if err := walker.err; err != nil {
		return err
	}

	return nil
}

type decryptionWalker struct {
	decrypter Decrypter
	err       error
}

func (w *decryptionWalker) Walk(node ast.Node) (newNode ast.Node, keepWalking bool) {
	// set the return values so we can have a simple return
	newNode = node
	keepWalking = true

	// look for an object item
	objectItem, ok := node.(*ast.ObjectItem)
	if !ok {
		return
	}

	// the single object item should have just one key
	if len(objectItem.Keys) != 1 {
		return
	}

	val, ok := objectItem.Val.(*ast.ObjectType)
	if !ok {
		return
	}
	if len(val.List.Items) != 1 {
		return
	}

	valItem := val.List.Items[0]

	// the single key should be an identifer and should be "ciphertext"
	keyToken := valItem.Keys[0].Token
	if keyToken.Type != token.IDENT && keyToken.Type != token.STRING {
		return
	}
	if keyToken.Text != "ciphertext" && keyToken.Text != `"ciphertext"` {
		return
	}

	valueLiteralType, ok := valItem.Val.(*ast.LiteralType)
	if !ok {
		return
	}

	if valueLiteralType.Token.Type != token.STRING && valueLiteralType.Token.Type != token.HEREDOC {
		return
	}

	// At this point we have an objectType that contains ciphertext to
	// decrypt.

	if w.decrypter == nil {
		w.err = errors.New("no encryption config present")
		keepWalking = false
		return
	}

	cipherText := valueLiteralType.Token.Text

	if valueLiteralType.Token.Type == token.STRING {
		cipherText = strings.TrimPrefix(cipherText, `"`)
		cipherText = strings.TrimSuffix(cipherText, `"`)
	} else if valueLiteralType.Token.Type == token.HEREDOC {
		// with a HEREDOC we know that it will have the format "<<XXX ... XXX"
		// so after the leading and trailing spaces are removed, we get the
		// content by removing the first word (<<XXX) and the last word (XXX).
		f := func(r rune) bool {
			return !unicode.IsSpace(r)
		}

		// trim leading/trailing spaces
		cipherText = strings.TrimSpace(cipherText)

		// remove first word (<<XXX)
		cipherText = strings.TrimLeftFunc(cipherText, f)

		// remove last word (XXX)
		cipherText = strings.TrimRightFunc(cipherText, f)
	}

	clearText, err := w.decrypter.DecryptString(cipherText)
	if err != nil {
		w.err = errors.Wrap(err).With(
			"line", valueLiteralType.Token.Pos.Line,
			"column", valueLiteralType.Token.Pos.Column,
		)
		keepWalking = false
		return
	}

	newVal := &ast.LiteralType{
		Token: token.Token{
			Pos:  valueLiteralType.Token.Pos,
			Type: token.STRING,
			JSON: valueLiteralType.Token.JSON,
			Text: clearText, // note that the ciphertext includes the quotes
		},
	}

	objectItem.Val = newVal
	objectItem.Assign = valueLiteralType.Token.Pos
	return
}
