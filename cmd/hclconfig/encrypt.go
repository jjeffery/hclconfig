package main

import (
	"io"
	"os"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/jjeffery/errors"
	"github.com/jjeffery/hclconfig/amzn"
	"github.com/jjeffery/hclconfig/astutil"
	"github.com/jjeffery/hclconfig/download"
)

func decryptFile(location string, inplace bool) error {
	d, err := download.Get(location)
	if err != nil {
		return err
	}
	if inplace && !d.IsLocal {
		return errors.New("cannot write to non-local config file").With(
			"location", location,
		)
	}
	file, err := hcl.ParseBytes(d.Body)
	if err != nil {
		return errors.Wrap(err).With(
			"location", location,
		)
	}
	decrypter, err := amzn.NewKey(file)
	if err != nil {
		return errors.Wrap(err).With(
			"location", location,
		)
	}
	if err = astutil.Decrypt(file, decrypter); err != nil {
		return err
	}
	if err := printNode(file, inplace, location); err != nil {
		return err
	}
	return nil
}

func encryptFile(location string, inplace bool, keywords []string, values []string) error {
	// get the file contents
	d, err := download.Get(location)
	if err != nil {
		return err
	}
	if inplace && !d.IsLocal {
		return errors.New("cannot write to non-local config file").With(
			"location", location,
		)
	}
	file, err := hcl.ParseBytes(d.Body)
	if err != nil {
		return errors.Wrap(err).With(
			"location", location,
		)
	}
	encrypter, err := amzn.NewKey(file)
	if err != nil {
		return errors.Wrap(err).With(
			"location", location,
		)
	}
	if err = astutil.Encrypt(file, encrypter, keywords, values); err != nil {
		return err
	}

	if err := printNode(file, inplace, location); err != nil {
		return err
	}
	return nil
}

func printNode(node ast.Node, inplace bool, location string) error {
	var out io.WriteCloser
	var err error
	if inplace {
		out, err = os.Create(location)
		if err != nil {
			return err
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	hclPrinter := printer.Config{
		SpacesWidth: 4,
	}

	if err = hclPrinter.Fprint(out, node); err != nil {
		return errors.Wrap(err, "cannot format HCL").With(
			"location", location,
		)
	}

	return nil
}
