package astcrypt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/jjeffery/errors"
)

// Encrypter is an interface for encrypting text.
type Encrypter interface {
	EncryptString(cleartext string) (ciphertext string, err error)
}

// Encrypt searches the AST for keys that match any of the keywords
// and values that match any of the values. Any key in the configuration
// file matches a keyword if it contains that keyword. Any value in the configuration
// file matches a valueword if it contains that valueword. Tests are case insensitive.
//
// When a match is detected it converts the form
//  key = "<unencrypted-data>"
// into the form
//	key {
//		ciphertext = "<encrypted-data>"
//	}
// The encrypter is used to encrypt cleartext. If the encrypter
// is nil this function will return success only if there is nothing
// in the AST to encrypt.
func Encrypt(node ast.Node, encrypter Encrypter, keywords []string, valuewords []string) error {
	// ensure that any modifications to keywords and values will not
	// modify the original slices
	if keywords != nil {
		keywords = keywords[:len(keywords):len(keywords)]
	}
	if valuewords != nil {
		valuewords = valuewords[:len(valuewords):len(valuewords)]
	}

	// add any more keywords from the config file for encrypting
	{
		var encryptionConfig struct {
			Encryption struct {
				Keywords []string
				Values   []string
			}
		}

		if err := hcl.DecodeObject(&encryptionConfig, node); err != nil {
			return errors.Wrap(err, "cannot decode encryption keywords")
		}
		keywords = append(keywords, encryptionConfig.Encryption.Keywords...)
		valuewords = append(valuewords, encryptionConfig.Encryption.Values...)
	}

	walker := encryptWalker{
		encrypter: encrypter,
		keywords:  make([]string, len(keywords)),
		values:    make([]string, len(valuewords)),
	}

	for i, keyword := range keywords {
		walker.keywords[i] = strings.TrimSpace(strings.ToLower(keyword))
	}
	for i, value := range valuewords {
		walker.values[i] = strings.TrimSpace(strings.ToLower(value))
	}

	ast.Walk(node, walker.Walk)
	if walker.err != nil {
		return walker.err
	}

	return nil
}

type encryptWalker struct {
	encrypter Encrypter
	keywords  []string
	values    []string
	err       error
}

func (w *encryptWalker) Walk(node ast.Node) (newNode ast.Node, keepWalking bool) {
	// set the return values so we can have a simple return
	newNode = node
	keepWalking = true

	objectItem, ok := node.(*ast.ObjectItem)
	if !ok {
		return
	}

	if len(objectItem.Keys) != 1 {
		return
	}

	val, ok := objectItem.Val.(*ast.LiteralType)
	if !ok {
		return
	}

	if val.Token.Type != token.STRING {
		// we only encrypt string types
		// TODO(jpj): should consider encrypting token.HEREDOC
		return
	}

	valueText := val.Token.Text
	keyText := objectItem.Keys[0].Token.Text

	if !containsAny(keyText, w.keywords) && !containsAny(valueText, w.values) {
		return
	}

	// At this point we have a match, which means that we will encrypt
	// the value. This requires constructing a new AST node for the encrypted
	// value. First, encrypt the value.

	if w.encrypter == nil {
		w.err = errors.New("no encryption config present")
		keepWalking = false
		return
	}

	ciphertext, err := w.encrypter.EncryptString(valueText)
	if err != nil {
		w.err = err
		keepWalking = false
		return
	}

	isJSON := val.Token.JSON
	newVal := &ast.ObjectType{
		Lbrace: val.Token.Pos,
		Rbrace: val.Token.Pos,
		List: &ast.ObjectList{
			Items: []*ast.ObjectItem{
				{
					Keys: []*ast.ObjectKey{
						{
							Token: token.Token{
								Type: token.IDENT,
								Text: "ciphertext",
								JSON: isJSON,
								Pos:  val.Token.Pos,
							},
						},
					},
					Assign: objectItem.Assign,
					Val: &ast.LiteralType{
						Token: token.Token{
							Type: token.STRING,
							Text: fmt.Sprintf("%q", ciphertext),
							JSON: isJSON,
							Pos:  val.Token.Pos,
						},
					},
				},
			},
		},
	}

	objectItem.Val = newVal
	objectItem.Assign = token.Pos{}
	return
}

// containsAny is used for determining whether s contains any of
// the substrings in substr. The values in substr are required to
// be in lower case. The value in s is converted to lower case before
// testing.
func containsAny(s string, substrs []string) bool {
	s = strings.ToLower(s)
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
