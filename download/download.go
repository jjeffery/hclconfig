// Package download knows how to download a file from
// an HTTP URL, an S3 URL or from the local filesystem.
package download

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jjeffery/errkind"

	"github.com/jjeffery/errors"
	"github.com/jjeffery/hclconfig/amzn"
)

// File represents a file that has been downloaded
// from HTTP, S3 or the local filesystem.
type File struct {
	Location     string
	Body         []byte
	ETag         string
	LastModified time.Time
	IsLocal      bool
}

// Head returns a file without the body. It can be used to determine
// if the file has changed.
func Head(location string) (*File, error) {
	// TODO(jpj): implement
	return nil, errkind.NotImplemented()
}

// Get returns a file from the specified location, including the body.
func Get(location string) (*File, error) {
	u, err := url.Parse(location)
	if err != nil {
		// not a valid URL, so treat as a local file
		return getLocal(location)
	}

	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return getHTTP(location)
	case "s3":
		bucket := u.Host
		key := strings.TrimPrefix(u.Path, "/")
		return getS3(location, bucket, key)
	case "file", "":
		return getLocal(u.Path)
	default:
		return nil, errors.New("cannot open file: unknown scheme").With(
			"location", location,
		)
	}
}

func getLocal(location string) (*File, error) {
	f, err := os.Open(location)
	if err != nil {
		// error message contains file name
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		// error message contains file name
		return nil, err
	}

	body, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read file").With(
			"location", location,
		)
	}

	return &File{
		Location:     location,
		Body:         body,
		LastModified: fi.ModTime(),
		IsLocal:      true,
	}, nil
}

func getHTTP(location string) (*File, error) {
	client := http.Client{
		Timeout: time.Minute,
	}
	request, err := http.NewRequest("GET", location, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create http request").With(
			"location", location,
		)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get file").With(
			"location", location,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("cannot get file").With(
			"location", location,
			"statusCode", response.StatusCode,
			"status", response.Status,
		)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read file body").With(
			"location", location,
		)
	}

	lastModified, _ := http.ParseTime(response.Header.Get("Last-Modified"))

	file := &File{
		Location:     location,
		Body:         body,
		ETag:         response.Header.Get("Etag"),
		LastModified: lastModified,
	}

	return file, nil
}

func getS3(location, bucket, key string) (*File, error) {
	etag, lastModified, body, err := amzn.Download(bucket, key)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, errors.Wrap(err, "cannot download from s3").With(
			"bucket", bucket,
			"key", key,
		)
	}

	file := &File{
		Location:     location,
		Body:         bodyBytes,
		ETag:         etag,
		LastModified: lastModified,
	}
	return file, nil
}
