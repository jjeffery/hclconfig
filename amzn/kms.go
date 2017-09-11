package amzn

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/jjeffery/hclconfig/encryption"
	"github.com/jjeffery/errors"
)

var (
	// encryptionContext is used for encrypting/decrypting the data
	// keys: its only purpose is to ensure that the ciphertext blob
	// is really intended for the purpose.
	encryptionContext = map[string]*string{
		"usage": aws.String("cryptconfig"),
	}
)

// NewKey creates a new data encryption key based on the contents
// of the configuration file.
func NewKey(node ast.Node) (encryption.Key, error) {
	var data struct {
		Encryption struct {
			KMS *string
		}
	}

	if err := hcl.DecodeObject(&data, node); err != nil {
		return nil, errors.Wrap(err, "cannot decode KMS encryption config")
	}

	if data.Encryption.KMS == nil {
		// no KMS config info
		return nil, nil
	}

	replacer := strings.NewReplacer("\n", "", "\r", "", "\t", "", " ", "")
	base64Blob := replacer.Replace(*data.Encryption.KMS)
	binaryBlob, err := base64.StdEncoding.DecodeString(base64Blob)
	if err != nil {
		return nil, errors.New("kms encryption: invalid dataKey: not base64")
	}

	kmssvc := kms.New(AWSSession())
	output, err := kmssvc.Decrypt(&kms.DecryptInput{
		CiphertextBlob:    binaryBlob,
		EncryptionContext: encryptionContext,
	})

	if err != nil {
		return nil, errors.Wrap(err, "kms encryption: cannot decrypt data key")
	}

	return encryption.Key(output.Plaintext), nil
}

// GenerateDataKey generates a new data encryption key that can
// be used in the configuration file.
func GenerateDataKey(keyID string) (dataKey string, keyARN string, err error) {
	kmssvc := kms.New(AWSSession())
	output, err := kmssvc.GenerateDataKey(&kms.GenerateDataKeyInput{
		KeyId:             aws.String(keyID),
		KeySpec:           aws.String("AES_256"),
		EncryptionContext: encryptionContext,
	})
	if err != nil {
		return "", "", errors.Wrap(err, "cannot generate data key").With(
			"keyID", keyID,
		)
	}

	dataKey = base64.StdEncoding.EncodeToString(output.CiphertextBlob)
	keyARN = *output.KeyId
	return dataKey, keyARN, nil
}
