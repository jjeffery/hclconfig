// Package download knows how to download a file from
// an HTTP URL, an S3 URL or from the local filesystem.
package download

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jjeffery/errors"
	"github.com/jjeffery/hclconfig/amzn"
)

var (
	httpClient = http.Client{
		Timeout: time.Minute,
	}
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
	return get(location, false)
}

// Get returns a file from the specified location, including the body.
func Get(location string) (*File, error) {
	return get(location, true)
}

func get(location string, includeBody bool) (*File, error) {
	u, err := url.Parse(location)
	if err != nil {
		// not a valid URL, so treat as a local file
		return getLocal(location, includeBody)
	}

	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return getHTTP(location, includeBody)
	case "s3":
		bucket := u.Host
		key := strings.TrimPrefix(u.Path, "/")
		return getS3(location, bucket, key, includeBody)
	case "file", "":
		return getLocal(u.Path, includeBody)
	default:
		return nil, errors.New("cannot open file: unknown scheme").With(
			"location", location,
		)
	}
}

func getLocal(location string, includeBody bool) (*File, error) {
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

	var body []byte

	if includeBody {
		body, err = ioutil.ReadAll(f)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read file").With(
				"location", location,
			)
		}
	}

	return &File{
		Location:     location,
		Body:         body,
		LastModified: fi.ModTime(),
		IsLocal:      true,
	}, nil
}

func getHTTP(location string, includeBody bool) (*File, error) {
	method := "GET"
	if !includeBody {
		method = "HEAD:"
	}
	request, err := http.NewRequest(method, location, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create http request").With(
			"location", location,
		)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get file").With(
			"location", location,
			"method", method,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("cannot get file").With(
			"location", location,
			"method", method,
			"statusCode", response.StatusCode,
			"status", response.Status,
		)
	}

	var body []byte

	if includeBody {
		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read file body").With(
				"location", location,
			)
		}
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

func getS3(location, bucket, key string, includeBody bool) (*File, error) {
	var etag string
	var lastModified time.Time
	var body io.ReadCloser
	var bodyBytes []byte
	var err error

	if includeBody {
		etag, lastModified, body, err = amzn.Get(bucket, key)
		if err != nil {
			return nil, err
		}
		defer body.Close()
		bodyBytes, err = ioutil.ReadAll(body)
		if err != nil {
			return nil, errors.Wrap(err, "cannot download from s3").With(
				"bucket", bucket,
				"key", key,
			)
		}
	} else {
		etag, lastModified, err = amzn.Head(bucket, key)
		if err != nil {
			return nil, err
		}
	}

	file := &File{
		Location:     location,
		Body:         bodyBytes,
		ETag:         etag,
		LastModified: lastModified,
	}
	return file, nil
}

func getS3Changed(bucket string, key string, etag string) (changed bool, err error) {
	return amzn.HasChanged(bucket, key, etag)
}
