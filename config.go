package hclconfig

import (
	"time"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/jjeffery/errors"
	"github.com/jjeffery/hclconfig/amzn"
	"github.com/jjeffery/hclconfig/astutil"
	"github.com/jjeffery/hclconfig/download"
)

// Get downloads the configuration file from the location, parses it
// and decrypts any sensitive data.
// The location can be a http/https URL, and S3 URL, or a local
// file path.
func Get(location string) (*File, error) {
	d, err := download.Get(location)
	if err != nil {
		return nil, err
	}
	node, err := hcl.ParseBytes(d.Body)
	if err != nil {
		return nil, errors.Wrap(err).With(
			"location", location,
		)
	}
	decrypter, err := amzn.NewKey(node)
	if err = astutil.Decrypt(node, decrypter); err != nil {
		return nil, errors.Wrap(err).With(
			"location", location,
		)
	}
	f := &File{
		Location:     location,
		Etag:         d.ETag,
		LastModified: d.LastModified,
		Contents:     node,
	}
	return f, nil
}

// File represents a configuration file that has been loaded
// from a location.
type File struct {
	Location     string
	Etag         string
	LastModified time.Time
	Contents     *ast.File
}

// HasChanged returns true if the config file has changed. It
// does not download the new contents.
//
// For HTTP(S) and S3 URLs, this function performs a HEAD operation
// and compares the ETag or the Last-Modified headers. For local files
// this function performs a file stat and compares the last modified times.
func (f *File) HasChanged() (bool, error) {
	d, err := download.Head(f.Location)
	if err != nil {
		return false, err
	}
	if d.ETag != "" && f.Etag != "" {
		// if both have ETags, the file has changed if they are not equal
		return d.ETag != f.Etag, nil
	}

	return d.LastModified.After(f.LastModified), nil
}

// Decode decodes the contents of the configuration file into the
// structure pointed to by v.
func (f *File) Decode(v interface{}) error {
	return hcl.DecodeObject(v, f.Contents)
}
